package pool

import (
	"context"
	"os/exec"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/events"
	"github.com/spiral/roadrunner/v2/ipc"
	"github.com/spiral/roadrunner/v2/payload"
	"github.com/spiral/roadrunner/v2/utils"
	"github.com/spiral/roadrunner/v2/worker"
	workerWatcher "github.com/spiral/roadrunner/v2/worker_watcher"
	"go.uber.org/zap"
)

const (
	// StopRequest can be sent by worker to indicate that restart is required.
	StopRequest = `{"stop":true}`
)

// ErrorEncoder encode error or make a decision based on the error type
type ErrorEncoder func(err error, w worker.BaseProcess) (*payload.Payload, error)

type Options func(p *StaticPool)

type Command func() *exec.Cmd

// StaticPool controls worker creation, destruction and task routing. Pool uses fixed amount of stack.
type StaticPool struct {
	cfg *Config
	log *zap.Logger

	// worker command creator
	cmd Command

	// creates and connects to stack
	factory ipc.Factory

	// manages worker states and TTLs
	ww Watcher

	// allocate new worker
	allocator worker.Allocator

	// errEncoder is the default Exec error encoder
	errEncoder ErrorEncoder
}

// NewStaticPool creates new worker pool and task multiplexer. StaticPool will initiate with one worker.
func NewStaticPool(ctx context.Context, cmd Command, factory ipc.Factory, cfg *Config, options ...Options) (Pool, error) {
	if factory == nil {
		return nil, errors.Str("no factory initialized")
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
	}

	// add pool options
	for i := 0; i < len(options); i++ {
		options[i](p)
	}

	if p.log == nil {
		z, err := zap.NewProduction()
		if err != nil {
			return nil, err
		}

		p.log = z
	}

	// set up workers allocator
	p.allocator = p.newPoolAllocator(ctx, p.cfg.AllocateTimeout, factory, cmd)
	// set up workers watcher
	p.ww = workerWatcher.NewSyncWorkerWatcher(p.allocator, p.log, p.cfg.NumWorkers, p.cfg.AllocateTimeout)

	// allocate requested number of workers
	workers, err := p.allocateWorkers(p.cfg.NumWorkers)
	if err != nil {
		return nil, err
	}

	// add workers to the watcher
	err = p.ww.Watch(workers)
	if err != nil {
		return nil, err
	}

	p.errEncoder = defaultErrEncoder(p)

	// if supervised config not nil, guess, that pool wanted to be supervised
	if cfg.Supervisor != nil {
		sp := supervisorWrapper(p, p.log, p.cfg.Supervisor)
		// start watcher timer
		sp.Start()
		return sp, nil
	}

	return p, nil
}

func WithLogger(z *zap.Logger) Options {
	return func(p *StaticPool) {
		p.log = z
	}
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
func (sp *StaticPool) Exec(p *payload.Payload) (*payload.Payload, error) {
	const op = errors.Op("static_pool_exec")
	if sp.cfg.Debug {
		return sp.execDebug(p)
	}
	ctxGetFree, cancel := context.WithTimeout(context.Background(), sp.cfg.AllocateTimeout)
	defer cancel()
	w, err := sp.takeWorker(ctxGetFree, op)
	if err != nil {
		return nil, errors.E(op, err)
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
	sp.ww.Release(w)
	return rsp, nil
}

// Destroy all underlying stack (but let them complete the task).
func (sp *StaticPool) Destroy(ctx context.Context) {
	sp.ww.Destroy(ctx)
}

func (sp *StaticPool) Reset(ctx context.Context) error {
	// destroy all workers
	sp.ww.Reset(ctx)
	workers, err := sp.allocateWorkers(sp.cfg.NumWorkers)
	if err != nil {
		return err
	}
	// add the NEW workers to the watcher
	err = sp.ww.Watch(workers)
	if err != nil {
		return err
	}

	return nil
}

func defaultErrEncoder(sp *StaticPool) ErrorEncoder {
	return func(err error, w worker.BaseProcess) (*payload.Payload, error) {
		// just push event if on any stage was timeout error
		switch {
		case errors.Is(errors.ExecTTL, err):
			sp.log.Warn("worker stopped, and will be restarted", zap.String("reason", "execTTL timeout elapsed"), zap.Int64("pid", w.Pid()), zap.String("internal_event_name", events.EventExecTTL.String()), zap.Error(err))
			w.State().Set(worker.StateInvalid)
			return nil, err

		case errors.Is(errors.SoftJob, err):
			sp.log.Warn("worker stopped, and will be restarted", zap.String("reason", "worker error"), zap.Int64("pid", w.Pid()), zap.String("internal_event_name", events.EventWorkerError.String()), zap.Error(err))
			// if max jobs exceed
			if sp.cfg.MaxJobs != 0 && w.State().NumExecs() >= sp.cfg.MaxJobs {
				// mark old as invalid and stop
				w.State().Set(worker.StateInvalid)
				errS := w.Stop()
				if errS != nil {
					return nil, errors.E(errors.SoftJob, errors.Errorf("err: %v\nerrStop: %v", err, errS))
				}

				return nil, err
			}

			// soft jobs errors are allowed, just put the worker back
			sp.ww.Release(w)

			return nil, err
		case errors.Is(errors.Network, err):
			// in case of network error, we can't stop the worker, we should kill it
			w.State().Set(worker.StateInvalid)
			sp.log.Warn("network error, worker will be restarted", zap.String("reason", "network"), zap.Int64("pid", w.Pid()), zap.String("internal_event_name", events.EventWorkerError.String()), zap.Error(err))
			// kill the worker instead of sending net packet to it
			_ = w.Kill()

			return nil, err
		default:
			w.State().Set(worker.StateInvalid)
			sp.log.Warn("worker will be restarted", zap.Int64("pid", w.Pid()), zap.String("internal_event_name", events.EventWorkerDestruct.String()), zap.Error(err))
			// stop the worker, worker here might be in the broken state (network)
			errS := w.Stop()
			if errS != nil {
				return nil, errors.E(errors.Errorf("err: %v\nerrStop: %v", err, errS))
			}

			return nil, err
		}
	}
}

// Be careful, sync with pool.Exec method
func (sp *StaticPool) execWithTTL(ctx context.Context, p *payload.Payload) (*payload.Payload, error) {
	const op = errors.Op("static_pool_exec_with_context")
	if sp.cfg.Debug {
		return sp.execDebugWithTTL(ctx, p)
	}

	ctxAlloc, cancel := context.WithTimeout(context.Background(), sp.cfg.AllocateTimeout)
	defer cancel()
	w, err := sp.takeWorker(ctxAlloc, op)
	if err != nil {
		return nil, errors.E(op, err)
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
	sp.ww.Release(w)
	return rsp, nil
}

func (sp *StaticPool) stopWorker(w worker.BaseProcess) {
	w.State().Set(worker.StateInvalid)
	err := w.Stop()
	if err != nil {
		sp.log.Warn("user requested worker to be stopped", zap.String("reason", "user event"), zap.Int64("pid", w.Pid()), zap.String("internal_event_name", events.EventWorkerError.String()), zap.Error(err))
	}
}

// checkMaxJobs check for worker number of executions and kill workers if that number more than sp.cfg.MaxJobs
func (sp *StaticPool) checkMaxJobs(w worker.BaseProcess) {
	if w.State().NumExecs() >= sp.cfg.MaxJobs {
		w.State().Set(worker.StateMaxJobsReached)
		sp.ww.Release(w)
		return
	}

	sp.ww.Release(w)
}

func (sp *StaticPool) takeWorker(ctxGetFree context.Context, op errors.Op) (worker.BaseProcess, error) {
	// Get function consumes context with timeout
	w, err := sp.ww.Take(ctxGetFree)
	if err != nil {
		// if the error is of kind NoFreeWorkers, it means, that we can't get worker from the stack during the allocate timeout
		if errors.Is(errors.NoFreeWorkers, err) {
			sp.log.Error("no free workers in the pool, wait timeout exceed", zap.String("reason", "no free workers"), zap.String("internal_event_name", events.EventNoFreeWorkers.String()), zap.Error(err))
			return nil, errors.E(op, err)
		}
		// else if err not nil - return error
		return nil, errors.E(op, err)
	}
	return w, nil
}

func (sp *StaticPool) newPoolAllocator(ctx context.Context, timeout time.Duration, factory ipc.Factory, cmd func() *exec.Cmd) worker.Allocator {
	return func() (worker.SyncWorker, error) {
		ctxT, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		w, err := factory.SpawnWorkerWithTimeout(ctxT, cmd())
		if err != nil {
			return nil, err
		}

		// wrap sync worker
		sw := worker.From(w)

		sp.log.Debug("worker is allocated", zap.Int64("pid", sw.Pid()), zap.String("internal_event_name", events.EventWorkerConstruct.String()))
		return sw, nil
	}
}

// execDebug used when debug mode was not set and exec_ttl is 0
func (sp *StaticPool) execDebug(p *payload.Payload) (*payload.Payload, error) {
	sw, err := sp.allocator()
	if err != nil {
		return nil, err
	}

	// redirect call to the workers' exec method (without ttl)
	r, err := sw.Exec(p)
	if err != nil {
		return nil, err
	}

	go func() {
		// read the exit status to prevent process to be a zombie
		_ = sw.Wait()
	}()

	// destroy the worker
	err = sw.Stop()
	if err != nil {
		sp.log.Debug("debug mode: worker stopped", zap.String("reason", "worker error"), zap.Int64("pid", sw.Pid()), zap.String("internal_event_name", events.EventWorkerError.String()), zap.Error(err))
		return nil, err
	}

	return r, nil
}

// execDebugWithTTL used when user set debug mode and exec_ttl
func (sp *StaticPool) execDebugWithTTL(ctx context.Context, p *payload.Payload) (*payload.Payload, error) {
	sw, err := sp.allocator()
	if err != nil {
		return nil, err
	}

	// redirect call to the worker with TTL
	r, err := sw.ExecWithTTL(ctx, p)
	if err != nil {
		return nil, err
	}

	go func() {
		// read the exit status to prevent process to be a zombie
		_ = sw.Wait()
	}()

	err = sw.Stop()
	if err != nil {
		sp.log.Debug("debug mode: worker stopped", zap.String("reason", "worker error"), zap.Int64("pid", sw.Pid()), zap.String("internal_event_name", events.EventWorkerError.String()), zap.Error(err))
		return nil, err
	}

	return r, err
}

// allocate required number of stack
func (sp *StaticPool) allocateWorkers(numWorkers uint64) ([]worker.BaseProcess, error) {
	workers := make([]worker.BaseProcess, 0, numWorkers)

	// constant number of stack simplify logic
	for i := uint64(0); i < numWorkers; i++ {
		w, err := sp.allocator()
		if err != nil {
			return nil, errors.E(errors.WorkerAllocate, err)
		}

		workers = append(workers, w)
	}
	return workers, nil
}
