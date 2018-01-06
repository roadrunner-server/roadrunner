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
	// Pid of the process, can be null if not started
	Pid *int

	st       State  // st current worker st.
	numExecs uint64 // numExecs show how many times worker have been invoked.

	cmd *exec.Cmd     // underlying command process
	err *bytes.Buffer // aggregates stderr

	mu sync.RWMutex  // ensures than only one execution can be run at once
	rl goridge.Relay // communication bus with underlying process
}

// newWorker creates new worker
func newWorker(cmd *exec.Cmd) (*Worker, error) {
	w := &Worker{
		cmd: cmd,
		err: bytes.NewBuffer(nil),
		st:  newState(StateDisabled),
	}

	if w.cmd.Process != nil {
		return nil, fmt.Errorf("can't attach to running process")
	}

	return w, nil
}

// State provides access to worker state
func (w *Worker) State() State {
	return w.st
}

// NumExecs show how many times worker have been invoked.
func (w *Worker) NumExecs() uint64 {
	return w.numExecs
}

// String returns worker description.
func (w *Worker) String() string {
	state := w.st.String()

	if w.Pid != nil {
		state = state + ", pid:" + strconv.Itoa(*w.Pid)
	}

	return fmt.Sprintf("(`%st` [%s], numExecs: %v)", strings.Join(w.cmd.Args, " "), state, w.numExecs)
}

// Start underlying process or return error
func (w *Worker) Start() error {
	stderr, err := w.cmd.StderrPipe()
	if err != nil {
		w.st = newState(StateError)
		return err
	}

	// copying all process errors into buffer space
	go io.Copy(w.err, stderr)

	if err := w.cmd.Start(); err != nil {
		w.st = newState(StateError)
		return w.mockError(err)
	}

	w.Pid = &w.cmd.Process.Pid

	return nil
}

// Stop underlying process (No timeout limit) or return error.
func (w *Worker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.st = newState(StateDisabled)

	go func() {
		sendCommand(w.rl, &stopCommand{Stop: true})
	}()

	w.cmd.Wait()
	w.rl.Close()

	w.st = newState(StateStopped)
}

// todo: timeout
// return syscall.Kill(-c.status.PID, syscall.SIGTERM)

// Exec command and return result and result context.
func (w *Worker) Exec(body []byte, ctx interface{}) (resp []byte, rCtx []byte, err error) {
	if w.st.Value() != StateReady {
		return nil, nil, fmt.Errorf("worker is not ready (%s)", w.st.Value())
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	defer atomic.AddUint64(&w.numExecs, 1)

	w.st = newState(StateWorking)

	if err := w.sendPayload(ctx, goridge.PayloadControl); err != nil {
		return nil, nil, fmt.Errorf("invalid context: %st", err)
	}

	w.rl.Send(body, 0)

	// response header
	rCtx, p, err := w.rl.Receive()

	if p.HasFlag(goridge.PayloadError) {
		w.st = newState(StateReady)
		return nil, nil, JobError(rCtx)
	}

	if !p.HasFlag(goridge.PayloadControl) {
		w.st = newState(StateError)
		return nil, nil, w.mockError(fmt.Errorf("invalid response (check script integrity)"))
	}

	// body
	if resp, p, err = w.rl.Receive(); err != nil {
		w.st = newState(StateError)
		return nil, nil, w.mockError(fmt.Errorf("worker error: %st", err))
	}

	w.st = newState(StateReady)
	return resp, rCtx, nil
}

// attach payload/control relay to the worker.
func (w *Worker) attach(rl goridge.Relay) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.rl = rl
	w.st = newState(StateAttached)
}

// mockError attaches worker specific error (from stderr) to parent error
func (w *Worker) mockError(err error) WorkerError {
	//w.cmd.Process.Signal(os.Kill)

	time.Sleep(time.Millisecond * 100)
	if w.err.Len() != 0 {
		return WorkerError(w.err.String())
	}

	return WorkerError(err.Error())
}

func (w *Worker) sendPayload(v interface{}, flags byte) error {
	if v == nil {
		w.rl.Send(nil, goridge.PayloadControl)
	}
	data, err := json.Marshal(v)

	if err != nil {
		return fmt.Errorf("invalid payload: %s", err)
	}

	return w.rl.Send(data, goridge.PayloadControl)
}
