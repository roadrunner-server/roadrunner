package roadrunner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spiral/goridge"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Worker - supervised process with api over goridge.Relay.
type Worker struct {
	// State current worker state.
	State State

	// Last time worker State has changed
	Last time.Time

	// NumExecutions how many times worker have been invoked.
	NumExecutions uint64

	// Pid contains process ID and empty until worker is started.
	Pid *int

	cmd *exec.Cmd     // underlying command process
	err *bytes.Buffer // aggregates stderr
	rl  goridge.Relay // communication bus with underlying process
	mu  sync.RWMutex  // ensures than only one execution can be run at once
}

// NewWorker creates new worker
func NewWorker(cmd *exec.Cmd) (*Worker, error) {
	w := &Worker{
		cmd:   cmd,
		err:   bytes.NewBuffer(nil),
		State: StateInactive,
	}

	if w.cmd.Process != nil {
		return nil, fmt.Errorf("can't attach to running process")
	}

	return w, nil
}

// String returns worker description.
func (w *Worker) String() string {
	state := w.State.String()

	if w.Pid != nil {
		state = state + ", pid:" + strconv.Itoa(*w.Pid)
	}

	return fmt.Sprintf("(`%s` [%s], execs: %v)", strings.Join(w.cmd.Args, " "), state, w.NumExecutions)
}

// Start underlying process or return error
func (w *Worker) Start() error {
	stderr, err := w.cmd.StderrPipe()
	if err != nil {
		w.setState(StateError)
		return err
	}

	// copying all process errors into buffer space
	go io.Copy(w.err, stderr)

	if err := w.cmd.Start(); err != nil {
		w.setState(StateError)
		return w.mockError(err)
	}

	w.setState(StateReady)

	return nil
}

// Execute command and return result and result context.
func (w *Worker) Execute(body []byte, ctx interface{}) (resp []byte, rCtx []byte, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.State != StateReady {
		return nil, nil, fmt.Errorf("worker must be in state `waiting` (`%s` given)", w.State)
	}

	w.setState(StateReady)
	atomic.AddUint64(&w.NumExecutions, 1)

	if ctx != nil {
		if data, err := json.Marshal(ctx); err == nil {
			w.rl.Send(data, goridge.PayloadControl)
		} else {
			return nil, nil, fmt.Errorf("invalid context: %s", err)
		}
	} else {
		w.rl.Send(nil, goridge.PayloadControl|goridge.PayloadEmpty)
	}

	w.rl.Send(body, 0)

	rCtx, p, err := w.rl.Receive()

	if !p.HasFlag(goridge.PayloadControl) {
		return nil, nil, w.mockError(fmt.Errorf("invalid response (check script integrity)"))
	}

	if p.HasFlag(goridge.PayloadError) {
		w.setState(StateReady)
		return nil, nil, JobError(rCtx)
	}

	if resp, p, err = w.rl.Receive(); err != nil {
		w.setState(StateError)
		return nil, nil, w.mockError(fmt.Errorf("worker error: %s", err))
	}

	w.setState(StateReady)
	return resp, rCtx, nil
}

// Stop underlying process or return error.
func (w *Worker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.setState(StateInactive)

	go func() {
		sendCommand(w.rl, &TerminateCommand{Terminate: true})
	}()

	w.cmd.Wait()
	w.rl.Close()

	w.setState(StateStopped)
}

// attach payload/control relay to the worker.
func (w *Worker) attach(rl goridge.Relay) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.rl = rl
	w.setState(StateBooting)
}

// sets worker State and it's context (non blocking!).
func (w *Worker) setState(state State) {
	// safer?
	w.State = state
	w.Last = time.Now()
}

// mockError attaches worker specific error (from stderr) to parent error
func (w *Worker) mockError(err error) WorkerError {
	if w.err.Len() != 0 {
		return WorkerError(w.err.String())
	}

	return WorkerError(err.Error())
}
