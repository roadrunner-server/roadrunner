package pipe

import (
	"context"
	"os/exec"

	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/pkg/pipe"
	"github.com/spiral/roadrunner/v2/interfaces/events"
	"github.com/spiral/roadrunner/v2/interfaces/worker"
	"github.com/spiral/roadrunner/v2/internal"
	workerImpl "github.com/spiral/roadrunner/v2/pkg/worker"
	"go.uber.org/multierr"
)

// Factory connects to stack using standard
// streams (STDIN, STDOUT pipes).
type Factory struct {
	listener []events.EventListener
}

// NewPipeFactory returns new factory instance and starts
// listening

// todo: review tests
func NewPipeFactory(listeners ...events.EventListener) worker.Factory {
	return &Factory{
		listener: listeners,
	}
}

type SpawnResult struct {
	w   worker.BaseProcess
	err error
}

// SpawnWorker creates new Process and connects it to goridge relay,
// method Wait() must be handled on level above.
func (f *Factory) SpawnWorkerWithTimeout(ctx context.Context, cmd *exec.Cmd) (worker.BaseProcess, error) {
	c := make(chan SpawnResult)
	const op = errors.Op("spawn worker with context")
	go func() {
		w, err := workerImpl.InitBaseWorker(cmd, workerImpl.AddListeners(f.listener...))
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
		relay := pipe.NewPipeRelay(in, out)
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
		pid, err := internal.FetchPID(relay)
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
		w.State().Set(internal.StateReady)

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

func (f *Factory) SpawnWorker(cmd *exec.Cmd) (worker.BaseProcess, error) {
	const op = errors.Op("spawn worker")
	w, err := workerImpl.InitBaseWorker(cmd, workerImpl.AddListeners(f.listener...))
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
	relay := pipe.NewPipeRelay(in, out)
	w.AttachRelay(relay)

	// Start the worker
	err = w.Start()
	if err != nil {
		return nil, errors.E(op, err)
	}

	// errors bundle
	if pid, err := internal.FetchPID(relay); pid != w.Pid() {
		err = multierr.Combine(
			err,
			w.Kill(),
			w.Wait(),
		)
		return nil, errors.E(op, err)
	}

	// everything ok, set ready state
	w.State().Set(internal.StateReady)
	return w, nil
}

// Close the factory.
func (f *Factory) Close() error {
	return nil
}
