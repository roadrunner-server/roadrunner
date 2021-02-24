package worker

import (
	"bytes"
	"context"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/pkg/frame"
	"github.com/spiral/goridge/v3/pkg/relay"
	"github.com/spiral/roadrunner/v2/pkg/payload"
	"go.uber.org/multierr"
)

// Allocator is responsible for worker allocation in the pool
type Allocator func() (SyncWorker, error)

type SyncWorkerImpl struct {
	process *Process
}

// From creates SyncWorker from BaseProcess
func From(process *Process) SyncWorker {
	return &SyncWorkerImpl{
		process: process,
	}
}

// Exec payload without TTL timeout.
func (tw *SyncWorkerImpl) Exec(p payload.Payload) (payload.Payload, error) {
	const op = errors.Op("sync_worker_exec")
	if len(p.Body) == 0 && len(p.Context) == 0 {
		return payload.Payload{}, errors.E(op, errors.Str("payload can not be empty"))
	}

	if tw.process.State().Value() != StateReady {
		return payload.Payload{}, errors.E(op, errors.Errorf("Process is not ready (%s)", tw.process.State().String()))
	}

	// set last used time
	tw.process.State().SetLastUsed(uint64(time.Now().UnixNano()))
	tw.process.State().Set(StateWorking)

	rsp, err := tw.execPayload(p)
	if err != nil {
		// just to be more verbose
		if errors.Is(errors.SoftJob, err) == false {
			tw.process.State().Set(StateErrored)
			tw.process.State().RegisterExec()
		}
		return payload.Payload{}, err
	}

	tw.process.State().Set(StateReady)
	tw.process.State().RegisterExec()

	return rsp, nil
}

type wexec struct {
	payload payload.Payload
	err     error
}

// Exec payload without TTL timeout.
func (tw *SyncWorkerImpl) ExecWithTTL(ctx context.Context, p payload.Payload) (payload.Payload, error) {
	const op = errors.Op("sync_worker_exec_worker_with_timeout")
	c := make(chan wexec, 1)

	go func() {
		if len(p.Body) == 0 && len(p.Context) == 0 {
			c <- wexec{
				payload: payload.Payload{},
				err:     errors.E(op, errors.Str("payload can not be empty")),
			}
			return
		}

		if tw.process.State().Value() != StateReady {
			c <- wexec{
				payload: payload.Payload{},
				err:     errors.E(op, errors.Errorf("Process is not ready (%s)", tw.process.State().String())),
			}
			return
		}

		// set last used time
		tw.process.State().SetLastUsed(uint64(time.Now().UnixNano()))
		tw.process.State().Set(StateWorking)

		rsp, err := tw.execPayload(p)
		if err != nil {
			// just to be more verbose
			if errors.Is(errors.SoftJob, err) == false {
				tw.process.State().Set(StateErrored)
				tw.process.State().RegisterExec()
			}
			c <- wexec{
				payload: payload.Payload{},
				err:     errors.E(op, err),
			}
			return
		}

		tw.process.State().Set(StateReady)
		tw.process.State().RegisterExec()

		c <- wexec{
			payload: rsp,
			err:     nil,
		}
	}()

	select {
	// exec TTL reached
	case <-ctx.Done():
		err := multierr.Combine(tw.Kill())
		if err != nil {
			// append timeout error
			err = multierr.Append(err, errors.E(op, errors.ExecTTL))
			return payload.Payload{}, multierr.Append(err, ctx.Err())
		}
		return payload.Payload{}, errors.E(op, errors.ExecTTL, ctx.Err())
	case res := <-c:
		if res.err != nil {
			return payload.Payload{}, res.err
		}
		return res.payload, nil
	}
}

func (tw *SyncWorkerImpl) execPayload(p payload.Payload) (payload.Payload, error) {
	const op = errors.Op("sync_worker_exec_payload")

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

	err = tw.process.Relay().Receive(frameR)
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
		return payload.Payload{}, errors.E(op, errors.SoftJob, errors.Str(string(frameR.Payload())))
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

func (tw *SyncWorkerImpl) String() string {
	return tw.process.String()
}

func (tw *SyncWorkerImpl) Pid() int64 {
	return tw.process.Pid()
}

func (tw *SyncWorkerImpl) Created() time.Time {
	return tw.process.Created()
}

func (tw *SyncWorkerImpl) State() State {
	return tw.process.State()
}

func (tw *SyncWorkerImpl) Start() error {
	return tw.process.Start()
}

func (tw *SyncWorkerImpl) Wait() error {
	return tw.process.Wait()
}

func (tw *SyncWorkerImpl) Stop() error {
	return tw.process.Stop()
}

func (tw *SyncWorkerImpl) Kill() error {
	return tw.process.Kill()
}

func (tw *SyncWorkerImpl) Relay() relay.Relay {
	return tw.process.Relay()
}

func (tw *SyncWorkerImpl) AttachRelay(rl relay.Relay) {
	tw.process.AttachRelay(rl)
}
