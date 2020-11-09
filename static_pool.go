package roadrunner

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/util"
)

// StopRequest can be sent by worker to indicate that restart is required.
const StopRequest = "{\"stop\":true}"

var bCtx = context.Background()

// StaticPool controls worker creation, destruction and task routing. Pool uses fixed amount of stack.
type StaticPool struct {
	cfg Config

	// worker command creator
	cmd func() *exec.Cmd

	// creates and connects to stack
	factory Factory

	// distributes the events
	events *util.EventHandler

	// manages worker states and TTLs
	ww *workerWatcher
}

// NewPool creates new worker pool and task multiplexer. StaticPool will initiate with one worker.
func NewPool(ctx context.Context, cmd func() *exec.Cmd, factory Factory, cfg Config) (Pool, error) {
	cfg.InitDefaults()

	if cfg.Debug {
		cfg.NumWorkers = 0
		cfg.MaxJobs = 1
	}

	p := &StaticPool{
		cfg:     cfg,
		cmd:     cmd,
		factory: factory,
		events:  &util.EventHandler{},
	}

	p.ww = newWorkerWatcher(func(args ...interface{}) (WorkerBase, error) {
		w, err := p.factory.SpawnWorkerWithContext(ctx, p.cmd())
		if err != nil {
			return nil, err
		}

		sw, err := NewSyncWorker(w)
		if err != nil {
			return nil, err
		}
		return sw, nil
	}, p.cfg.NumWorkers, p.events)

	workers, err := p.allocateWorkers(ctx, p.cfg.NumWorkers)
	if err != nil {
		return nil, err
	}

	// put stack in the pool
	err = p.ww.AddToWatch(ctx, workers)
	if err != nil {
		return nil, err
	}

	// if supervised config not nil, guess, that pool wanted to be supervised
	if cfg.Supervisor != nil {
		sp := newPoolWatcher(p, p.events, p.cfg.Supervisor)
		// start watcher timer
		sp.Start()
		return sp, nil
	}

	return p, nil
}

// AddListener connects event listener to the pool.
func (sp *StaticPool) AddListener(listener util.EventListener) {
	sp.events.AddListener(listener)
}

// Config returns associated pool configuration. Immutable.
func (sp *StaticPool) GetConfig() Config {
	return sp.cfg
}

// Workers returns worker list associated with the pool.
func (sp *StaticPool) Workers() (workers []WorkerBase) {
	return sp.ww.WorkersList()
}

func (sp *StaticPool) RemoveWorker(ctx context.Context, wb WorkerBase) error {
	return sp.ww.RemoveWorker(ctx, wb)
}

func (sp *StaticPool) Exec(p Payload) (Payload, error) {
	const op = errors.Op("Exec")
	if sp.cfg.Debug {
		return sp.execDebug(p)
	}
	w, err := sp.ww.GetFreeWorker(context.Background())
	if err != nil && errors.Is(errors.ErrWatcherStopped, err) {
		return EmptyPayload, errors.E(op, err)
	} else if err != nil {
		return EmptyPayload, err
	}

	sw := w.(SyncWorker)

	rsp, err := sw.Exec(p)
	if err != nil {
		// soft job errors are allowed
		if errors.Is(errors.Exec, err) {
			if sp.cfg.MaxJobs != 0 && w.State().NumExecs() >= sp.cfg.MaxJobs {
				err = sp.ww.AllocateNew(bCtx)
				if err != nil {
					sp.events.Push(PoolEvent{Event: EventPoolError, Payload: err})
				}

				w.State().Set(StateInvalid)
				err = w.Stop(bCtx)
				if err != nil {
					sp.events.Push(WorkerEvent{Event: EventWorkerError, Worker: w, Payload: err})
				}
			} else {
				sp.ww.PushWorker(w)
			}

			return EmptyPayload, err

		}

		sw.State().Set(StateInvalid)
		sp.events.Push(PoolEvent{Event: EventWorkerDestruct, Payload: w})
		errS := w.Stop(bCtx)

		if errS != nil {
			return EmptyPayload, fmt.Errorf("%v, %v", err, errS)
		}

		return EmptyPayload, err
	}

	// worker want's to be terminated
	if rsp.Body == nil && rsp.Context != nil && string(rsp.Context) == StopRequest {
		w.State().Set(StateInvalid)
		err = w.Stop(bCtx)
		if err != nil {
			sp.events.Push(WorkerEvent{Event: EventWorkerError, Worker: w, Payload: err})
		}

		return sp.Exec(p)
	}

	if sp.cfg.MaxJobs != 0 && w.State().NumExecs() >= sp.cfg.MaxJobs {
		err = sp.ww.AllocateNew(bCtx)
		if err != nil {
			return EmptyPayload, err
		}
	} else {
		sp.ww.PushWorker(w)
	}
	return rsp, nil
}

func (sp *StaticPool) execDebug(p Payload) (Payload, error) {
	sw, err := sp.ww.allocator()
	if err != nil {
		return EmptyPayload, err
	}

	r, err := sw.(SyncWorker).Exec(p)

	if stopErr := sw.Stop(context.Background()); stopErr != nil {
		sp.events.Push(WorkerEvent{Event: EventWorkerError, Worker: sw, Payload: err})
	}

	return r, err
}

func (sp *StaticPool) ExecWithContext(ctx context.Context, rqs Payload) (Payload, error) {
	const op = errors.Op("Exec")
	w, err := sp.ww.GetFreeWorker(context.Background())
	if err != nil && errors.Is(errors.ErrWatcherStopped, err) {
		return EmptyPayload, errors.E(op, err)
	} else if err != nil {
		return EmptyPayload, err
	}

	sw := w.(SyncWorker)

	rsp, err := sw.ExecWithContext(ctx, rqs)
	if err != nil {
		// soft job errors are allowed
		if errors.Is(errors.Exec, err) {
			if sp.cfg.MaxJobs != 0 && w.State().NumExecs() >= sp.cfg.MaxJobs {
				err = sp.ww.AllocateNew(bCtx)
				if err != nil {
					sp.events.Push(PoolEvent{Event: EventPoolError, Payload: err})
				}

				w.State().Set(StateInvalid)
				err = w.Stop(bCtx)
				if err != nil {
					sp.events.Push(WorkerEvent{Event: EventWorkerError, Worker: w, Payload: err})
				}
			} else {
				sp.ww.PushWorker(w)
			}

			return EmptyPayload, err
		}

		sw.State().Set(StateInvalid)
		sp.events.Push(PoolEvent{Event: EventWorkerDestruct, Payload: w})
		errS := w.Stop(bCtx)

		if errS != nil {
			return EmptyPayload, fmt.Errorf("%v, %v", err, errS)
		}

		return EmptyPayload, err
	}

	// worker want's to be terminated
	if rsp.Body == nil && rsp.Context != nil && string(rsp.Context) == StopRequest {
		w.State().Set(StateInvalid)
		err = w.Stop(bCtx)
		if err != nil {
			sp.events.Push(WorkerEvent{Event: EventWorkerError, Worker: w, Payload: err})
		}

		return sp.Exec(rqs)
	}

	if sp.cfg.MaxJobs != 0 && w.State().NumExecs() >= sp.cfg.MaxJobs {
		err = sp.ww.AllocateNew(bCtx)
		if err != nil {
			return EmptyPayload, err
		}
	} else {
		sp.ww.PushWorker(w)
	}
	return rsp, nil
}

// Destroy all underlying stack (but let them to complete the task).
func (sp *StaticPool) Destroy(ctx context.Context) {
	sp.ww.Destroy(ctx)
}

// allocate required number of stack
func (sp *StaticPool) allocateWorkers(ctx context.Context, numWorkers int64) ([]WorkerBase, error) {
	var workers []WorkerBase

	// constant number of stack simplify logic
	for i := int64(0); i < numWorkers; i++ {
		ctx, cancel := context.WithTimeout(ctx, sp.cfg.AllocateTimeout)
		w, err := sp.factory.SpawnWorkerWithContext(ctx, sp.cmd())
		if err != nil {
			cancel()
			return nil, err
		}
		cancel()
		workers = append(workers, w)
	}
	return workers, nil
}

func (sp *StaticPool) checkMaxJobs(ctx context.Context, w WorkerBase) error {
	if sp.cfg.MaxJobs != 0 && w.State().NumExecs() >= sp.cfg.MaxJobs {
		err := sp.ww.AllocateNew(ctx)
		if err != nil {
			sp.events.Push(PoolEvent{Event: EventPoolError, Payload: err})
			return err
		}
	}
	return nil
}
