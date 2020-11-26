package roadrunner

import (
	"context"
	"os/exec"

	"github.com/spiral/errors"
	"github.com/spiral/goridge/v2"
	"go.uber.org/multierr"
)

// PipeFactory connects to stack using standard
// streams (STDIN, STDOUT pipes).
type PipeFactory struct{}

// NewPipeFactory returns new factory instance and starts
// listening

// todo: review tests
func NewPipeFactory() Factory {
	return &PipeFactory{}
}

type SpawnResult struct {
	w   WorkerBase
	err error
}

// SpawnWorker creates new WorkerProcess and connects it to goridge relay,
// method Wait() must be handled on level above.
func (f *PipeFactory) SpawnWorkerWithContext(ctx context.Context, cmd *exec.Cmd) (WorkerBase, error) {
	c := make(chan SpawnResult)
	const op = errors.Op("spawn worker with context")
	go func() {
		w, err := InitBaseWorker(cmd)
		if err != nil {
			c <- SpawnResult{
				w:   nil,
				err: errors.E(op, err),
			}
			return
		}

		// TODO why out is in?
		in, err := cmd.StdoutPipe()
		if err != nil {
			c <- SpawnResult{
				w:   nil,
				err: errors.E(op, err),
			}
			return
		}

		// TODO why in is out?
		out, err := cmd.StdinPipe()
		if err != nil {
			c <- SpawnResult{
				w:   nil,
				err: errors.E(op, err),
			}
			return
		}

		// Init new PIPE relay
		relay := goridge.NewPipeRelay(in, out)
		w.AttachRelay(relay)

		// Start the worker
		err = w.Start()
		if err != nil {
			c <- SpawnResult{
				w:   nil,
				err: errors.E(op, err),
			}
			return
		}

		// errors bundle
		pid, err := fetchPID(relay)
		if pid != w.Pid() || err != nil {
			err = multierr.Combine(
				err,
				w.Kill(),
				w.Wait(),
			)
			c <- SpawnResult{
				w:   nil,
				err: errors.E(op, err),
			}
			return
		}

		// everything ok, set ready state
		w.State().Set(StateReady)

		// return worker
		c <- SpawnResult{
			w:   w,
			err: nil,
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-c:
		if res.err != nil {
			return nil, res.err
		}
		return res.w, nil
	}
}

func (f *PipeFactory) SpawnWorker(cmd *exec.Cmd) (WorkerBase, error) {
	const op = errors.Op("spawn worker")
	w, err := InitBaseWorker(cmd)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// TODO why out is in?
	in, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.E(op, err)
	}

	// TODO why in is out?
	out, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.E(op, err)
	}

	// Init new PIPE relay
	relay := goridge.NewPipeRelay(in, out)
	w.AttachRelay(relay)

	// Start the worker
	err = w.Start()
	if err != nil {
		return nil, errors.E(op, err)
	}

	// errors bundle
	if pid, err := fetchPID(relay); pid != w.Pid() {
		err = multierr.Combine(
			err,
			w.Kill(),
			w.Wait(),
		)
		return nil, errors.E(op, err)
	}

	// everything ok, set ready state
	w.State().Set(StateReady)
	return w, nil
}

// Close the factory.
func (f *PipeFactory) Close(ctx context.Context) error {
	return nil
}
