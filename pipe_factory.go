package roadrunner

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/spiral/goridge/v2"
)

// PipeFactory connects to workers using standard
// streams (STDIN, STDOUT pipes).
type PipeFactory struct {
}

// NewPipeFactory returns new factory instance and starts
// listening
func NewPipeFactory() *PipeFactory {
	return &PipeFactory{}
}

// SpawnWorker creates new worker and connects it to goridge relay,
// method Wait() must be handled on level above.
func (f *PipeFactory) SpawnWorker(cmd *exec.Cmd) (w *Worker, err error) {
	if w, err = newWorker(cmd); err != nil {
		return nil, err
	}

	var (
		in  io.ReadCloser
		out io.WriteCloser
	)

	if in, err = cmd.StdoutPipe(); err != nil {
		return nil, err
	}

	if out, err = cmd.StdinPipe(); err != nil {
		return nil, err
	}

	w.rl = goridge.NewPipeRelay(in, out)

	if err := w.start(); err != nil {
		return nil, errors.Wrap(err, "process error")
	}

	if pid, err := fetchPID(w.rl); pid != *w.Pid {
		go func(w *Worker) {
			err := w.Kill()
			if err != nil {
				// there is no logger here, how to handle error in goroutines ?
				fmt.Printf("error killing the worker with PID number %d, Created: %s", w.Pid, w.Created)
			}
		}(w)

		if wErr := w.Wait(); wErr != nil {
			if _, ok := wErr.(*exec.ExitError); ok {
				// error might be nil here
				if err != nil {
					err = errors.Wrap(wErr, err.Error())
				}
			} else {
				err = wErr
			}
		}

		return nil, errors.Wrap(err, "unable to connect to worker")
	}

	w.state.set(StateReady)
	return w, nil
}

// Close the factory.
func (f *PipeFactory) Close() error {
	return nil
}
