package roadrunner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/spiral/goridge/v2"
)

// PipeFactory connects to stack using standard
// streams (STDIN, STDOUT pipes).
type PipeFactory struct {
}

// NewPipeFactory returns new factory instance and starts
// listening

// todo: review tests
func NewPipeFactory() *PipeFactory {
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
	go func() {
		w, err := InitBaseWorker(cmd)
		if err != nil {
			c <- SpawnResult{
				w:   nil,
				err: err,
			}
			return
		}

		// TODO why out is in?
		in, err := cmd.StdoutPipe()
		if err != nil {
			c <- SpawnResult{
				w:   nil,
				err: err,
			}
			return
		}

		// TODO why in is out?
		out, err := cmd.StdinPipe()
		if err != nil {
			c <- SpawnResult{
				w:   nil,
				err: err,
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
				err: errors.Wrap(err, "process error"),
			}
			return
		}

		// errors bundle
		var errs []string
		if pid, errF := fetchPID(relay); pid != w.Pid() {
			if errF != nil {
				errs = append(errs, errF.Error())
			}

			// todo kill timeout
			errK := w.Kill(ctx)
			if errK != nil {
				errs = append(errs, fmt.Errorf("error killing the worker with PID number %d, Created: %s", w.Pid(), w.Created()).Error())
			}

			if wErr := w.Wait(ctx); wErr != nil {
				errs = append(errs, wErr.Error())
			}

			if len(errs) > 0 {
				c <- SpawnResult{
					w:   nil,
					err: errors.New(strings.Join(errs, " : ")),
				}
			} else {
				c <- SpawnResult{
					w:   nil,
					err: nil,
				}
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
	w, err := InitBaseWorker(cmd)
	if err != nil {
		return nil, err
	}

	// TODO why out is in?
	in, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// TODO why in is out?
	out, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	// Init new PIPE relay
	relay := goridge.NewPipeRelay(in, out)
	w.AttachRelay(relay)

	// Start the worker
	err = w.Start()
	if err != nil {
		return nil, errors.Wrap(err, "process error")
	}

	// errors bundle
	var errs []string
	if pid, errF := fetchPID(relay); pid != w.Pid() {
		if errF != nil {
			errs = append(errs, errF.Error())
		}

		// todo kill timeout ??
		errK := w.Kill(context.Background())
		if errK != nil {
			errs = append(errs, fmt.Errorf("error killing the worker with PID number %d, Created: %s", w.Pid(), w.Created()).Error())
		}

		if wErr := w.Wait(context.Background()); wErr != nil {
			errs = append(errs, wErr.Error())
		}

		if len(errs) > 0 {
			return nil, errors.New(strings.Join(errs, "/"))
		}
	}

	// everything ok, set ready state
	w.State().Set(StateReady)
	return w, nil
}

// Close the factory.
func (f *PipeFactory) Close(ctx context.Context) error {
	return nil
}
