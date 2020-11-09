package roadrunner

import (
	"context"
	"fmt"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/util"
	"go.uber.org/multierr"

	"github.com/spiral/goridge/v2"
)

var EmptyPayload = Payload{}

type SyncWorker interface {
	// WorkerBase provides basic functionality for the SyncWorker
	WorkerBase

	// Exec used to execute payload on the SyncWorker, there is no TIMEOUTS
	Exec(rqs Payload) (Payload, error)
	// ExecWithContext used to handle Exec with TTL
	ExecWithContext(ctx context.Context, p Payload) (Payload, error)
}

type syncWorker struct {
	w WorkerBase
}

func NewSyncWorker(w WorkerBase) (SyncWorker, error) {
	return &syncWorker{
		w: w,
	}, nil
}

// Exec payload without TTL timeout.
func (tw *syncWorker) Exec(p Payload) (Payload, error) {
	if len(p.Body) == 0 && len(p.Context) == 0 {
		return EmptyPayload, fmt.Errorf("payload can not be empty")
	}

	if tw.w.State().Value() != StateReady {
		return EmptyPayload, fmt.Errorf("WorkerProcess is not ready (%s)", tw.w.State().String())
	}

	// set last used time
	tw.w.State().SetLastUsed(uint64(time.Now().UnixNano()))
	tw.w.State().Set(StateWorking)

	rsp, err := tw.execPayload(p)
	if err != nil {
		// just to be more verbose
		if errors.Is(errors.Exec, err) == false {
			tw.w.State().Set(StateErrored)
			tw.w.State().RegisterExec()
		}
		return EmptyPayload, err
	}

	tw.w.State().Set(StateReady)
	tw.w.State().RegisterExec()

	return rsp, nil
}

type wexec struct {
	payload Payload
	err     error
}

// Exec payload without TTL timeout.
func (tw *syncWorker) ExecWithContext(ctx context.Context, p Payload) (Payload, error) {
	const op = errors.Op("exec_with_context")
	c := make(chan wexec, 1)
	go func() {
		if len(p.Body) == 0 && len(p.Context) == 0 {
			c <- wexec{
				payload: EmptyPayload,
				err:     errors.E(op, errors.Str("payload can not be empty")),
			}
			return
		}

		if tw.w.State().Value() != StateReady {
			c <- wexec{
				payload: EmptyPayload,
				err:     errors.E(op, errors.Errorf("WorkerProcess is not ready (%s)", tw.w.State().String())),
			}
			return
		}

		// set last used time
		tw.w.State().SetLastUsed(uint64(time.Now().UnixNano()))
		tw.w.State().Set(StateWorking)

		rsp, err := tw.execPayload(p)
		if err != nil {
			// just to be more verbose
			if errors.Is(errors.Exec, err) == false {
				tw.w.State().Set(StateErrored)
				tw.w.State().RegisterExec()
			}
			c <- wexec{
				payload: EmptyPayload,
				err:     errors.E(op, err),
			}
			return
		}

		tw.w.State().Set(StateReady)
		tw.w.State().RegisterExec()

		c <- wexec{
			payload: rsp,
			err:     nil,
		}
	}()

	select {
	case <-ctx.Done():
		err := multierr.Combine(tw.Kill())
		if err != nil {
			return EmptyPayload, multierr.Append(err, ctx.Err())
		}
		return EmptyPayload, ctx.Err()
	case res := <-c:
		if res.err != nil {
			return EmptyPayload, res.err
		}
		return res.payload, nil
	}
}

func (tw *syncWorker) execPayload(p Payload) (Payload, error) {
	const op = errors.Op("exec_payload")
	// two things; todo: merge
	if err := sendControl(tw.w.Relay(), p.Context); err != nil {
		return EmptyPayload, errors.E(op, err, "header error")
	}

	if err := tw.w.Relay().Send(p.Body, 0); err != nil {
		return EmptyPayload, errors.E(op, err, "sender error")
	}

	var pr goridge.Prefix
	rsp := Payload{}

	var err error
	if rsp.Context, pr, err = tw.w.Relay().Receive(); err != nil {
		return EmptyPayload, errors.E(op, err, "WorkerProcess error")
	}

	if !pr.HasFlag(goridge.PayloadControl) {
		return EmptyPayload, fmt.Errorf("malformed WorkerProcess response")
	}

	if pr.HasFlag(goridge.PayloadError) {
		return EmptyPayload, errors.E(op, errors.Exec, errors.Str(string(rsp.Context))) //ExecError(rsp.Context)
	}

	// add streaming support :)
	if rsp.Body, pr, err = tw.w.Relay().Receive(); err != nil {
		return EmptyPayload, errors.E(op, err, "WorkerProcess error")
	}

	return rsp, nil
}

func (tw *syncWorker) String() string {
	return tw.w.String()
}

func (tw *syncWorker) Pid() int64 {
	return tw.w.Pid()
}

func (tw *syncWorker) Created() time.Time {
	return tw.w.Created()
}

func (tw *syncWorker) AddListener(listener util.EventListener) {
	tw.w.AddListener(listener)
}

func (tw *syncWorker) State() State {
	return tw.w.State()
}

func (tw *syncWorker) Start() error {
	return tw.w.Start()
}

func (tw *syncWorker) Wait(ctx context.Context) error {
	return tw.w.Wait(ctx)
}

func (tw *syncWorker) Stop(ctx context.Context) error {
	return tw.w.Stop(ctx)
}

func (tw *syncWorker) Kill() error {
	return tw.w.Kill()
}

func (tw *syncWorker) Relay() goridge.Relay {
	return tw.w.Relay()
}

func (tw *syncWorker) AttachRelay(rl goridge.Relay) {
	tw.w.AttachRelay(rl)
}
