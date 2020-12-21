package worker

import (
	"bytes"
	"context"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/interfaces/relay"
	"github.com/spiral/goridge/v3/pkg/frame"
	"github.com/spiral/roadrunner/v2/interfaces/events"
	"github.com/spiral/roadrunner/v2/interfaces/worker"
	"github.com/spiral/roadrunner/v2/internal"
	"github.com/spiral/roadrunner/v2/pkg/payload"
	"go.uber.org/multierr"
)

type syncWorker struct {
	w worker.BaseProcess
}

// From creates SyncWorker from WorkerBasa
func From(w worker.BaseProcess) (worker.SyncWorker, error) {
	return &syncWorker{
		w: w,
	}, nil
}

// Exec payload without TTL timeout.
func (tw *syncWorker) Exec(p payload.Payload) (payload.Payload, error) {
	const op = errors.Op("sync worker Exec")
	if len(p.Body) == 0 && len(p.Context) == 0 {
		return payload.Payload{}, errors.E(op, errors.Str("payload can not be empty"))
	}

	if tw.w.State().Value() != internal.StateReady {
		return payload.Payload{}, errors.E(op, errors.Errorf("Process is not ready (%s)", tw.w.State().String()))
	}

	// set last used time
	tw.w.State().SetLastUsed(uint64(time.Now().UnixNano()))
	tw.w.State().Set(internal.StateWorking)

	rsp, err := tw.execPayload(p)
	if err != nil {
		// just to be more verbose
		if errors.Is(errors.ErrSoftJob, err) == false {
			tw.w.State().Set(internal.StateErrored)
			tw.w.State().RegisterExec()
		}
		return payload.Payload{}, err
	}

	tw.w.State().Set(internal.StateReady)
	tw.w.State().RegisterExec()

	return rsp, nil
}

type wexec struct {
	payload payload.Payload
	err     error
}

// Exec payload without TTL timeout.
func (tw *syncWorker) ExecWithTimeout(ctx context.Context, p payload.Payload) (payload.Payload, error) {
	const op = errors.Op("ExecWithTimeout")
	c := make(chan wexec, 1)

	go func() {
		if len(p.Body) == 0 && len(p.Context) == 0 {
			c <- wexec{
				payload: payload.Payload{},
				err:     errors.E(op, errors.Str("payload can not be empty")),
			}
			return
		}

		if tw.w.State().Value() != internal.StateReady {
			c <- wexec{
				payload: payload.Payload{},
				err:     errors.E(op, errors.Errorf("Process is not ready (%s)", tw.w.State().String())),
			}
			return
		}

		// set last used time
		tw.w.State().SetLastUsed(uint64(time.Now().UnixNano()))
		tw.w.State().Set(internal.StateWorking)

		rsp, err := tw.execPayload(p)
		if err != nil {
			// just to be more verbose
			if errors.Is(errors.ErrSoftJob, err) == false {
				tw.w.State().Set(internal.StateErrored)
				tw.w.State().RegisterExec()
			}
			c <- wexec{
				payload: payload.Payload{},
				err:     errors.E(op, err),
			}
			return
		}

		tw.w.State().Set(internal.StateReady)
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
			return payload.Payload{}, multierr.Append(err, ctx.Err())
		}
		return payload.Payload{}, ctx.Err()
	case res := <-c:
		if res.err != nil {
			return payload.Payload{}, res.err
		}
		return res.payload, nil
	}
}

func (tw *syncWorker) execPayload(p payload.Payload) (payload.Payload, error) {
	const op = errors.Op("exec pl")

	fr := frame.NewFrame()
	fr.WriteVersion(frame.VERSION_1)
	// can be 0 here

	buf := new(bytes.Buffer)
	buf.Write(p.Context)
	buf.Write(p.Body)

	// Context offset
	fr.WriteOptions(uint32(len(p.Context)))
	fr.WritePayloadLen(uint32(buf.Len()))
	fr.WritePayload(buf.Bytes())

	fr.WriteCRC()

	// empty and free the buffer
	buf.Truncate(0)

	err := tw.Relay().Send(fr)
	if err != nil {
		return payload.Payload{}, err
	}

	frameR := frame.NewFrame()

	err = tw.w.Relay().Receive(frameR)
	if err != nil {
		return payload.Payload{}, errors.E(op, err)
	}
	if frameR == nil {
		return payload.Payload{}, errors.E(op, errors.Str("nil fr received"))
	}

	if !frameR.VerifyCRC() {
		return payload.Payload{}, errors.E(op, errors.Str("failed to verify CRC"))
	}

	flags := frameR.ReadFlags()

	if flags&byte(frame.ERROR) != byte(0) {
		return payload.Payload{}, errors.E(op, errors.ErrSoftJob, errors.Str(string(frameR.Payload())))
	}

	options := frameR.ReadOptions()
	if len(options) != 1 {
		return payload.Payload{}, errors.E(op, errors.Str("options length should be equal 1 (body offset)"))
	}

	pl := payload.Payload{}
	pl.Context = frameR.Payload()[:options[0]]
	pl.Body = frameR.Payload()[options[0]:]

	return pl, nil
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

func (tw *syncWorker) AddListener(listener events.EventListener) {
	tw.w.AddListener(listener)
}

func (tw *syncWorker) State() internal.State {
	return tw.w.State()
}

func (tw *syncWorker) Start() error {
	return tw.w.Start()
}

func (tw *syncWorker) Wait() error {
	return tw.w.Wait()
}

func (tw *syncWorker) Stop() error {
	return tw.w.Stop()
}

func (tw *syncWorker) Kill() error {
	return tw.w.Kill()
}

func (tw *syncWorker) Relay() relay.Relay {
	return tw.w.Relay()
}

func (tw *syncWorker) AttachRelay(rl relay.Relay) {
	tw.w.AttachRelay(rl)
}
