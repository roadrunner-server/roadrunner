package pool

import (
	"context"
	"os/exec"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/payload"
	"github.com/spiral/roadrunner/v2/pkg/transport"
	"github.com/spiral/roadrunner/v2/pkg/worker"
	workerWatcher "github.com/spiral/roadrunner/v2/pkg/worker_watcher"
	"github.com/spiral/roadrunner/v2/utils"
)

// StopRequest can be sent by worker to indicate that restart is required.
const StopRequest = "{\"stop\":true}"

// ErrorEncoder encode error or make a decision based on the error type
type ErrorEncoder func(err error, w worker.BaseProcess) (payload.Payload, error)

type Options func(p *StaticPool)

type Command func() *exec.Cmd

// StaticPool controls worker creation, destruction and task routing. Pool uses fixed amount of stack.
type StaticPool struct {
	cfg Config

	// worker command creator
	cmd Command

	// creates and connects to stack
	factory transport.Factory

	// distributes the events
	events events.Handler

	// saved list of event listeners
	listeners []events.Listener

	// manages worker states and TTLs
	ww Watcher

	// allocate new worker
	allocator worker.Allocator

	// errEncoder is the default Exec error encoder
	errEncoder ErrorEncoder
}

// Initialize creates new worker pool and task multiplexer. StaticPool will initiate with one worker.
func Initialize(ctx context.Context, cmd Command, factory transport.Factory, cfg Config, options ...Options) (Pool, error) {
	const op = errors.Op("static_pool_initialize")
	if factory == nil {
		return nil, errors.E(op, errors.Str("no factory initialized"))
	}
	cfg.InitDefaults()

	if cfg.Debug {
		cfg.NumWorkers = 0
		cfg.MaxJobs = 1
	}

	p := &StaticPool{
		cfg:     cfg,
		cmd:     cmd,
		factory: factory,
		events:  events.NewEventsHandler(),
	}

	// add pool options
	for i := 0; i < len(options); i++ {
		options[i](p)
	}

	// set up workers allocator
	p.allocator = p.newPoolAllocator(ctx, p.cfg.AllocateTimeout, factory, cmd)
	// set up workers watcher
	p.ww = workerWatcher.NewSyncWorkerWatcher(p.allocator, p.cfg.NumWorkers, p.events)

	// allocate requested number of workers
	workers, err := p.allocateWorkers(p.cfg.NumWorkers)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// add workers to the watcher
	err = p.ww.Watch(workers)
	if err != nil {
		return nil, errors.E(op, err)
	}

	p.errEncoder = defaultErrEncoder(p)

	// if supervised config not nil, guess, that pool wanted to be supervised
	if cfg.Supervisor != nil {
		sp := supervisorWrapper(p, p.events, p.cfg.Supervisor)
		// start watcher timer
		sp.Start()
		return sp, nil
	}

	return p, nil
}

func AddListeners(listeners ...events.Listener) Options {
	return func(p *StaticPool) {
		p.listeners = listeners
		for i := 0; i < len(listeners); i++ {
			p.addListener(listeners[i])
		}
	}
}

// AddListener connects event listener to the pool.
func (sp *StaticPool) addListener(listener events.Listener) {
	sp.events.AddListener(listener)
}

// GetConfig returns associated pool configuration. Immutable.
func (sp *StaticPool) GetConfig() interface{} {
	return sp.cfg
}

// Workers returns worker list associated with the pool.
func (sp *StaticPool) Workers() (workers []worker.BaseProcess) {
	return sp.ww.List()
}

func (sp *StaticPool) RemoveWorker(wb worker.BaseProcess) error {
	sp.ww.Remove(wb)
	return nil
}

// Exec executes provided payload on the worker
func (sp *StaticPool) Exec(p payload.Payload) (payload.Payload, error) {
	const op = errors.Op("static_pool_exec")
	if sp.cfg.Debug {
		return sp.execDebug(p)
	}
	ctxGetFree, cancel := context.WithTimeout(context.Background(), sp.cfg.AllocateTimeout)
	defer cancel()
	w, err := sp.getWorker(ctxGetFree, op)
	if err != nil {
		return payload.Payload{}, errors.E(op, err)
	}

	rsp, err := w.(worker.SyncWorker).Exec(p)
	if err != nil {
		return sp.errEncoder(err, w)
	}

	// worker want's to be terminated
	if len(rsp.Body) == 0 && utils.AsString(rsp.Context) == StopRequest {
		sp.stopWorker(w)
		return sp.Exec(p)
	}

	if sp.cfg.MaxJobs != 0 {
		sp.checkMaxJobs(w)
		return rsp, nil
	}
	// return worker back
	sp.ww.Push(w)
	return rsp, nil
}

// Be careful, sync with pool.Exec method
func (sp *StaticPool) execWithTTL(ctx context.Context, p payload.Payload) (payload.Payload, error) {
	const op = errors.Op("static_pool_exec_with_context")
	if sp.cfg.Debug {
		return sp.execDebugWithTTL(ctx, p)
	}

	ctxAlloc, cancel := context.WithTimeout(context.Background(), sp.cfg.AllocateTimeout)
	defer cancel()
	w, err := sp.getWorker(ctxAlloc, op)
	if err != nil {
		return payload.Payload{}, errors.E(op, err)
	}

	rsp, err := w.(worker.SyncWorker).ExecWithTTL(ctx, p)
	if err != nil {
		return sp.errEncoder(err, w)
	}

	// worker want's to be terminated
	if len(rsp.Body) == 0 && utils.AsString(rsp.Context) == StopRequest {
		sp.stopWorker(w)
		return sp.execWithTTL(ctx, p)
	}

	if sp.cfg.MaxJobs != 0 {
		sp.checkMaxJobs(w)
		return rsp, nil
	}

	// return worker back
	sp.ww.Push(w)
	return rsp, nil
}

func (sp *StaticPool) stopWorker(w worker.BaseProcess) {
	const op = errors.Op("static_pool_stop_worker")
	w.State().Set(worker.StateInvalid)
	err := w.Stop()
	if err != nil {
		sp.events.Push(events.WorkerEvent{Event: events.EventWorkerError, Worker: w, Payload: errors.E(op, err)})
	}
}

// checkMaxJobs check for worker number of executions and kill workers if that number more than sp.cfg.MaxJobs
//go:inline
func (sp *StaticPool) checkMaxJobs(w worker.BaseProcess) {
	if w.State().NumExecs() >= sp.cfg.MaxJobs {
		w.State().Set(worker.StateMaxJobsReached)
		sp.ww.Push(w)
		return
	}

	sp.ww.Push(w)
}

func (sp *StaticPool) getWorker(ctxGetFree context.Context, op errors.Op) (worker.BaseProcess, error) {
	// Get function consumes context with timeout
	w, err := sp.ww.Get(ctxGetFree)
	if err != nil {
		// if the error is of kind NoFreeWorkers, it means, that we can't get worker from the stack during the allocate timeout
		if errors.Is(errors.NoFreeWorkers, err) {
			sp.events.Push(events.PoolEvent{Event: events.EventNoFreeWorkers, Payload: errors.E(op, err)})
			return nil, errors.E(op, err)
		}
		// else if err not nil - return error
		return nil, errors.E(op, err)
	}
	return w, nil
}

// Destroy all underlying stack (but let them to complete the task).
func (sp *StaticPool) Destroy(ctx context.Context) {
	sp.ww.Destroy(ctx)
}

func defaultErrEncoder(sp *StaticPool) ErrorEncoder {
	return func(err error, w worker.BaseProcess) (payload.Payload, error) {
		const op = errors.Op("error_encoder")
		// just push event if on any stage was timeout error
		switch {
		case errors.Is(errors.ExecTTL, err):
			sp.events.Push(events.PoolEvent{Event: events.EventExecTTL, Payload: errors.E(op, err)})

		case errors.Is(errors.SoftJob, err):
			if sp.cfg.MaxJobs != 0 && w.State().NumExecs() >= sp.cfg.MaxJobs {
				err = sp.ww.Allocate()
				if err != nil {
					sp.events.Push(events.PoolEvent{Event: events.EventWorkerConstruct, Payload: errors.E(op, err)})
				}

				w.State().Set(worker.StateInvalid)
				err = w.Stop()
				if err != nil {
					sp.events.Push(events.WorkerEvent{Event: events.EventWorkerError, Worker: w, Payload: errors.E(op, err)})
				}
			} else {
				sp.events.Push(events.WorkerEvent{Event: events.EventWorkerError, Worker: w, Payload: err})
				sp.ww.Push(w)
			}
		}

		w.State().Set(worker.StateInvalid)
		sp.events.Push(events.PoolEvent{Event: events.EventWorkerDestruct, Payload: w})
		errS := w.Stop()
		if errS != nil {
			return payload.Payload{}, errors.E(op, err, errS)
		}

		return payload.Payload{}, errors.E(op, err)
	}
}

func (sp *StaticPool) newPoolAllocator(ctx context.Context, timeout time.Duration, factory transport.Factory, cmd func() *exec.Cmd) worker.Allocator {
	return func() (worker.SyncWorker, error) {
		ctxT, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		w, err := factory.SpawnWorkerWithTimeout(ctxT, cmd(), sp.listeners...)
		if err != nil {
			return nil, err
		}

		sw := worker.From(w)

		sp.events.Push(events.PoolEvent{
			Event:   events.EventWorkerConstruct,
			Payload: sw,
		})
		return sw, nil
	}
}

// execDebug used when debug mode was not set and exec_ttl is 0
func (sp *StaticPool) execDebug(p payload.Payload) (payload.Payload, error) {
	sw, err := sp.allocator()
	if err != nil {
		return payload.Payload{}, err
	}

	// redirect call to the workers exec method (without ttl)
	r, err := sw.Exec(p)
	if stopErr := sw.Stop(); stopErr != nil {
		sp.events.Push(events.WorkerEvent{Event: events.EventWorkerError, Worker: sw, Payload: err})
	}

	return r, err
}

// execDebugWithTTL used when user set debug mode and exec_ttl
func (sp *StaticPool) execDebugWithTTL(ctx context.Context, p payload.Payload) (payload.Payload, error) {
	sw, err := sp.allocator()
	if err != nil {
		return payload.Payload{}, err
	}

	// redirect call to the worker with TTL
	r, err := sw.ExecWithTTL(ctx, p)
	if stopErr := sw.Stop(); stopErr != nil {
		sp.events.Push(events.WorkerEvent{Event: events.EventWorkerError, Worker: sw, Payload: err})
	}

	return r, err
}

// allocate required number of stack
func (sp *StaticPool) allocateWorkers(numWorkers uint64) ([]worker.BaseProcess, error) {
	const op = errors.Op("allocate workers")
	workers := make([]worker.BaseProcess, 0, numWorkers)

	// constant number of stack simplify logic
	for i := uint64(0); i < numWorkers; i++ {
		w, err := sp.allocator()
		if err != nil {
			return nil, errors.E(op, errors.WorkerAllocate, err)
		}

		workers = append(workers, w)
	}
	return workers, nil
}
