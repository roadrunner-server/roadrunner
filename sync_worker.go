package roadrunner

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spiral/goridge/v2"
)

var EmptyPayload = Payload{}

type SyncWorker interface {
	// WorkerBase provides basic functionality for the SyncWorker
	WorkerBase
	// Exec used to execute payload on the SyncWorker, there is no TIMEOUTS
	Exec(rqs Payload) (Payload, error)

	// ExecWithContext allow to set ExecTTL
	ExecWithContext(ctx context.Context, rqs Payload) (Payload, error)
}

type taskWorker struct {
	w WorkerBase
}

func NewSyncWorker(w WorkerBase) (SyncWorker, error) {
	return &taskWorker{
		w: w,
	}, nil
}

type twexec struct {
	payload Payload
	err     error
}

func (tw *taskWorker) ExecWithContext(ctx context.Context, rqs Payload) (Payload, error) {
	c := make(chan twexec)
	go func() {
		if len(rqs.Body) == 0 && len(rqs.Context) == 0 {
			c <- twexec{
				payload: EmptyPayload,
				err:     fmt.Errorf("payload can not be empty"),
			}
			return
		}

		if tw.w.State().Value() != StateReady {
			c <- twexec{
				payload: EmptyPayload,
				err:     fmt.Errorf("WorkerProcess is not ready (%s)", tw.w.State().String()),
			}
			return
		}

		// set last used time
		tw.w.State().SetLastUsed(uint64(time.Now().UnixNano()))
		tw.w.State().Set(StateWorking)

		rsp, err := tw.execPayload(rqs)
		if err != nil {
			if _, ok := err.(TaskError); !ok {
				tw.w.State().Set(StateErrored)
				tw.w.State().RegisterExec()
			}
			c <- twexec{
				payload: EmptyPayload,
				err:     err,
			}
			return
		}

		tw.w.State().Set(StateReady)
		tw.w.State().RegisterExec()
		c <- twexec{
			payload: rsp,
			err:     nil,
		}
		return
	}()

	for {
		select {
		case <-ctx.Done():
			return EmptyPayload, ctx.Err()
		case res := <-c:
			if res.err != nil {
				return EmptyPayload, res.err
			}

			return res.payload, nil
		}
	}
}

//
func (tw *taskWorker) Exec(rqs Payload) (Payload, error) {
	if len(rqs.Body) == 0 && len(rqs.Context) == 0 {
		return EmptyPayload, fmt.Errorf("payload can not be empty")
	}

	if tw.w.State().Value() != StateReady {
		return EmptyPayload, fmt.Errorf("WorkerProcess is not ready (%s)", tw.w.State().String())
	}

	// set last used time
	tw.w.State().SetLastUsed(uint64(time.Now().UnixNano()))
	tw.w.State().Set(StateWorking)

	rsp, err := tw.execPayload(rqs)
	if err != nil {
		if _, ok := err.(TaskError); !ok {
			tw.w.State().Set(StateErrored)
			tw.w.State().RegisterExec()
		}
		return EmptyPayload, err
	}

	tw.w.State().Set(StateReady)
	tw.w.State().RegisterExec()

	return rsp, nil
}

func (tw *taskWorker) execPayload(rqs Payload) (Payload, error) {
	// two things; todo: merge
	if err := sendControl(tw.w.Relay(), rqs.Context); err != nil {
		return EmptyPayload, errors.Wrap(err, "header error")
	}

	if err := tw.w.Relay().Send(rqs.Body, 0); err != nil {
		return EmptyPayload, errors.Wrap(err, "sender error")
	}

	var pr goridge.Prefix
	rsp := Payload{}

	var err error
	if rsp.Context, pr, err = tw.w.Relay().Receive(); err != nil {
		return EmptyPayload, errors.Wrap(err, "WorkerProcess error")
	}

	if !pr.HasFlag(goridge.PayloadControl) {
		return EmptyPayload, fmt.Errorf("malformed WorkerProcess response")
	}

	if pr.HasFlag(goridge.PayloadError) {
		return EmptyPayload, TaskError(rsp.Context)
	}

	// add streaming support :)
	if rsp.Body, pr, err = tw.w.Relay().Receive(); err != nil {
		return EmptyPayload, errors.Wrap(err, "WorkerProcess error")
	}

	return rsp, nil
}

func (tw *taskWorker) String() string {
	return tw.w.String()
}

func (tw *taskWorker) Created() time.Time {
	return tw.w.Created()
}

func (tw *taskWorker) Events() <-chan WorkerEvent {
	return tw.w.Events()
}

func (tw *taskWorker) Pid() int64 {
	return tw.w.Pid()
}

func (tw *taskWorker) State() State {
	return tw.w.State()
}

func (tw *taskWorker) Start() error {
	return tw.w.Start()
}

func (tw *taskWorker) Wait(ctx context.Context) error {
	return tw.w.Wait(ctx)
}

func (tw *taskWorker) Stop(ctx context.Context) error {
	return tw.w.Stop(ctx)
}

func (tw *taskWorker) Kill(ctx context.Context) error {
	return tw.w.Kill(ctx)
}

func (tw *taskWorker) Relay() goridge.Relay {
	return tw.w.Relay()
}

func (tw *taskWorker) AttachRelay(rl goridge.Relay) {
	tw.w.AttachRelay(rl)
}
