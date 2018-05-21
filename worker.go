package roadrunner

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spiral/goridge"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// Worker - supervised process with api over goridge.Relay.
type Worker struct {
	// Pid of the process, points to Pid of underlying process and
	// can be nil while process is not started.
	Pid *int

	// state holds information about current worker state,
	// number of worker executions, last status change time.
	// publicly this object is receive-only and protected using Mutex
	// and atomic counter.
	state *state

	// underlying command with associated process, command must be
	// provided to worker from outside in non-started form. Cmd
	// stdErr direction will be handled by worker to aggregate error message.
	cmd *exec.Cmd

	// err aggregates stderr output from underlying process. Value can be
	// receive only once command is completed and all pipes are closed.
	err *bytes.Buffer

	// channel is being closed once command is complete.
	waitDone chan interface{}

	// contains information about resulted process state.
	endState *os.ProcessState

	// ensures than only one execution can be run at once.
	mu sync.Mutex

	// communication bus with underlying process.
	rl goridge.Relay
}

// newWorker creates new worker over given exec.cmd.
func newWorker(cmd *exec.Cmd) (*Worker, error) {
	if cmd.Process != nil {
		return nil, fmt.Errorf("can't attach to running process")
	}

	w := &Worker{
		cmd:      cmd,
		err:      new(bytes.Buffer),
		waitDone: make(chan interface{}),
		state:    newState(StateInactive),
	}

	// piping all stderr to command buffer
	w.cmd.Stderr = w.err

	return w, nil
}

// State return receive-only worker state object, state can be used to safely access
// worker status, time when status changed and number of worker executions.
func (w *Worker) State() State {
	return w.state
}

// String returns worker description.
func (w *Worker) String() string {
	state := w.state.String()
	if w.Pid != nil {
		state = state + ", pid:" + strconv.Itoa(*w.Pid)
	}

	return fmt.Sprintf(
		"(`%s` [%s], numExecs: %v)",
		strings.Join(w.cmd.Args, " "),
		state,
		w.state.NumExecs(),
	)
}

// Wait must be called once for each worker, call will be released once worker is
// complete and will return process error (if any), if stderr is presented it's value
// will be wrapped as WorkerError. Method will return error code if php process fails
// to find or start the script.
func (w *Worker) Wait() error {
	<-w.waitDone

	// ensure that all receive/send operations are complete
	w.mu.Lock()
	defer w.mu.Unlock()

	if runtime.GOOS != "windows" {
		// windows handles processes and close pipes differently,
		// we can ignore wait here as process.Wait() already being handled above
		w.cmd.Wait()
	}

	if w.endState.Success() {
		return nil
	}

	if w.err.Len() != 0 {
		return errors.New(w.err.String())
	}

	// generic process error
	return &exec.ExitError{ProcessState: w.endState}
}

// Stop sends soft termination command to the worker and waits for process completion.
func (w *Worker) Stop() error {
	select {
	case <-w.waitDone:
		return nil
	default:
		w.mu.Lock()
		defer w.mu.Unlock()

		w.state.set(StateInactive)
		err := sendPayload(w.rl, &stopCommand{Stop: true})

		<-w.waitDone
		return err
	}
}

// Kill kills underlying process, make sure to call Wait() func to gather
// error log from the stderr. Waits for process completion.
func (w *Worker) Kill() error {
	select {
	case <-w.waitDone:
		return nil
	default:
		w.mu.Lock()
		defer w.mu.Unlock()

		w.state.set(StateInactive)
		err := w.cmd.Process.Signal(os.Kill)

		<-w.waitDone
		return err
	}
}

// Exec sends payload to worker, executes it and returns result or
// error. Make sure to handle worker.Wait() to gather worker level
// errors. Method might return JobError indicating issue with payload.
func (w *Worker) Exec(rqs *Payload) (rsp *Payload, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if rqs == nil {
		return nil, fmt.Errorf("payload can not be empty")
	}

	if w.state.Value() != StateReady {
		return nil, fmt.Errorf("worker is not ready (%s)", w.state.String())
	}

	w.state.set(StateWorking)
	defer w.state.registerExec()

	rsp, err = w.execPayload(rqs)

	if err != nil {
		if _, ok := err.(JobError); !ok {
			w.state.set(StateErrored)
			return nil, err
		}
	}

	w.state.set(StateReady)
	return rsp, err
}

func (w *Worker) start() error {
	if err := w.cmd.Start(); err != nil {
		close(w.waitDone)
		return err
	}

	w.Pid = &w.cmd.Process.Pid

	// wait for process to complete
	go func() {
		w.endState, _ = w.cmd.Process.Wait()
		if w.waitDone != nil {
			w.state.set(StateStopped)
			close(w.waitDone)

			if w.rl != nil {
				w.mu.Lock()
				defer w.mu.Unlock()
				w.rl.Close()
			}
		}
	}()

	return nil
}

func (w *Worker) execPayload(rqs *Payload) (rsp *Payload, err error) {
	if err := sendPayload(w.rl, rqs.Context); err != nil {
		return nil, errors.Wrap(err, "header error")
	}

	w.rl.Send(rqs.Body, 0)

	var pr goridge.Prefix
	rsp = new(Payload)

	if rsp.Context, pr, err = w.rl.Receive(); err != nil {
		return nil, errors.Wrap(err, "worker error")
	}

	if !pr.HasFlag(goridge.PayloadControl) {
		return nil, fmt.Errorf("mailformed worker response")
	}

	if pr.HasFlag(goridge.PayloadError) {
		return nil, JobError(rsp.Context)
	}

	if rsp.Body, pr, err = w.rl.Receive(); err != nil {
		return nil, errors.Wrap(err, "worker error")
	}

	return rsp, nil
}
