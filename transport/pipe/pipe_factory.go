package pipe

import (
	"context"
	"os/exec"

	"github.com/spiral/goridge/v3/pkg/pipe"
	"github.com/spiral/roadrunner/v2/internal"
	"github.com/spiral/roadrunner/v2/worker"
)

// Factory connects to stack using standard
// streams (STDIN, STDOUT pipes).
type Factory struct{}

// NewPipeFactory returns new factory instance and starts
// listening
func NewPipeFactory() *Factory {
	return &Factory{}
}

type sr struct {
	w   *worker.Process
	err error
}

// SpawnWorkerWithTimeout creates new Process and connects it to goridge relay,
// method Wait() must be handled on level above.
func (f *Factory) SpawnWorkerWithTimeout(ctx context.Context, cmd *exec.Cmd) (*worker.Process, error) {
	spCh := make(chan sr)
	go func() {
		w, err := worker.InitBaseWorker(cmd)
		if err != nil {
			select {
			case spCh <- sr{
				w:   nil,
				err: err,
			}:
				return
			default:
				return
			}
		}

		in, err := cmd.StdoutPipe()
		if err != nil {
			select {
			case spCh <- sr{
				w:   nil,
				err: err,
			}:
				return
			default:
				return
			}
		}

		out, err := cmd.StdinPipe()
		if err != nil {
			select {
			case spCh <- sr{
				w:   nil,
				err: err,
			}:
				return
			default:
				return
			}
		}

		// Init new PIPE relay
		relay := pipe.NewPipeRelay(in, out)
		w.AttachRelay(relay)

		// Start the worker
		err = w.Start()
		if err != nil {
			select {
			case spCh <- sr{
				w:   nil,
				err: err,
			}:
				return
			default:
				return
			}
		}

		// used as a ping
		_, err = internal.Pid(relay)
		if err != nil {
			_ = w.Kill()
			select {
			case spCh <- sr{
				w:   nil,
				err: err,
			}:
				return
			default:
				_ = w.Kill()
				return
			}
		}

		select {
		case
		// return worker
		spCh <- sr{
			w:   w,
			err: nil,
		}:
			// everything ok, set ready state
			w.State().Set(worker.StateReady)
			return
		default:
			_ = w.Kill()
			return
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-spCh:
		if res.err != nil {
			return nil, res.err
		}
		return res.w, nil
	}
}

func (f *Factory) SpawnWorker(cmd *exec.Cmd) (*worker.Process, error) {
	w, err := worker.InitBaseWorker(cmd)
	if err != nil {
		return nil, err
	}

	in, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	out, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	// Init new PIPE relay
	relay := pipe.NewPipeRelay(in, out)
	w.AttachRelay(relay)

	// Start the worker
	err = w.Start()
	if err != nil {
		return nil, err
	}

	// errors bundle
	_, err = internal.Pid(relay)
	if err != nil {
		_ = w.Kill()
		return nil, err
	}

	// everything ok, set ready state
	w.State().Set(worker.StateReady)
	return w, nil
}

// Close the factory.
func (f *Factory) Close() error {
	return nil
}
