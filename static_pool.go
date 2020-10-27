package roadrunner

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	"github.com/spiral/roadrunner/v2/util"

	"github.com/pkg/errors"
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

	// protects state of worker list, does not affect allocation
	muw sync.RWMutex

	// manages worker states and TTLs
	ww *workerWatcher

	// supervises memory and TTL of workers
	// sp *supervisedPool
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

	// todo: implement
	// p.sp = newPoolWatcher(p, p.events, p.cfg.Supervisor)
	// p.sp.Start()

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
	if sp.cfg.Debug {
		return sp.execDebug(p)
	}

	w, err := sp.ww.GetFreeWorker(context.Background())
	if err != nil && errors.Is(err, ErrWatcherStopped) {
		return EmptyPayload, ErrWatcherStopped
	} else if err != nil {
		return EmptyPayload, err
	}

	sw := w.(SyncWorker)

	rsp, err := sw.Exec(p)
	if err != nil {
		// soft job errors are allowed
		if _, jobError := err.(ExecError); jobError {
			if sp.cfg.MaxJobs != 0 && w.State().NumExecs() >= sp.cfg.MaxJobs {
				err := sp.ww.AllocateNew(bCtx)
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

// Exec one task with given payload and context, returns result or error.
// func (p *StaticPool) ExecWithContext(ctx context.Context, rqs Payload) (Payload, error) {
//	// todo: why TODO passed here?
//	getWorkerCtx, cancel := context.WithTimeout(context.TODO(), p.cfg.AllocateTimeout)
//	defer cancel()
//	w, err := p.ww.GetFreeWorker(getWorkerCtx)
//	if err != nil && errors.Is(err, ErrWatcherStopped) {
//		return EmptyPayload, ErrWatcherStopped
//	} else if err != nil {
//		return EmptyPayload, err
//	}
//
//	sw := w.(SyncWorker)
//
//	// todo: implement worker destroy
//	//execCtx context.Context
//	//if p.cfg.Supervisor.ExecTTL != 0 {
//	//	var cancel2 context.CancelFunc
//	//	execCtx, cancel2 = context.WithTimeout(context.TODO(), p.cfg.Supervisor.ExecTTL)
//	//	defer cancel2()
//	//} else {
//	//	execCtx = ctx
//	//}
//
//	rsp, err := sw.Exec(rqs)
//	if err != nil {
//		errJ := p.checkMaxJobs(ctx, w)
//		if errJ != nil {
//			// todo: worker was not destroyed
//			return EmptyPayload, fmt.Errorf("%v, %v", err, errJ)
//		}
//
//		// soft job errors are allowed
//		if _, jobError := err.(ExecError); jobError {
//			p.ww.PushWorker(w)
//			return EmptyPayload, err
//		}
//
//		sw.State().Set(StateInvalid)
//		errS := w.Stop(ctx)
//		if errS != nil {
//			return EmptyPayload, fmt.Errorf("%v, %v", err, errS)
//		}
//
//		return EmptyPayload, err
//	}
//
//	// worker want's to be terminated
//	if rsp.Body == nil && rsp.Context != nil && string(rsp.Context) == StopRequest {
//		w.State().Set(StateInvalid)
//		err = w.Stop(ctx)
//		if err != nil {
//			return EmptyPayload, err
//		}
//		return p.ExecWithContext(ctx, rqs)
//	}
//
//	if p.cfg.MaxJobs != 0 && w.State().NumExecs() >= p.cfg.MaxJobs {
//		err = p.ww.AllocateNew(ctx)
//		if err != nil {
//			return EmptyPayload, err
//		}
//	} else {
//		p.muw.Lock()
//		p.ww.PushWorker(w)
//		p.muw.Unlock()
//	}
//
//	return rsp, nil
// }

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
