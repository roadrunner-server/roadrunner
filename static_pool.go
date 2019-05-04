package roadrunner

import (
	"fmt"
	"github.com/pkg/errors"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// StopRequest can be sent by worker to indicate that restart is required.
	StopRequest = "{\"stop\":true}"
)

// StaticPool controls worker creation, destruction and task routing. Pool uses fixed amount of workers.
type StaticPool struct {
	// pool behaviour
	cfg Config

	// worker command creator
	cmd func() *exec.Cmd

	// creates and connects to workers
	factory Factory

	// active task executions
	tmu   sync.Mutex
	tasks sync.WaitGroup

	// workers circular allocation buf
	free chan *Worker

	// number of workers expected to be dead in a buf.
	numDead int64

	// protects state of worker list, does not affect allocation
	muw sync.RWMutex

	// all registered workers
	workers []*Worker

	// invalid declares set of workers to be removed from the pool.
	mur    sync.Mutex
	remove sync.Map

	// pool is being destroyed
	inDestroy int32
	destroy   chan interface{}

	// lsn is optional callback to handle worker create/destruct/error events.
	mul sync.Mutex
	lsn func(event int, ctx interface{})
}

// NewPool creates new worker pool and task multiplexer. StaticPool will initiate with one worker.
func NewPool(cmd func() *exec.Cmd, factory Factory, cfg Config) (*StaticPool, error) {
	if err := cfg.Valid(); err != nil {
		return nil, errors.Wrap(err, "config")
	}

	p := &StaticPool{
		cfg:     cfg,
		cmd:     cmd,
		factory: factory,
		workers: make([]*Worker, 0, cfg.NumWorkers),
		free:    make(chan *Worker, cfg.NumWorkers),
		destroy: make(chan interface{}),
	}

	// constant number of workers simplify logic
	for i := int64(0); i < p.cfg.NumWorkers; i++ {
		// to test if worker ready
		w, err := p.createWorker()
		if err != nil {
			p.Destroy()
			return nil, err
		}

		p.free <- w
	}

	return p, nil
}

// Listen attaches pool event controller.
func (p *StaticPool) Listen(l func(event int, ctx interface{})) {
	p.mul.Lock()
	defer p.mul.Unlock()

	p.lsn = l

	p.muw.Lock()
	for _, w := range p.workers {
		w.err.Listen(p.lsn)
	}
	p.muw.Unlock()
}

// Config returns associated pool configuration. Immutable.
func (p *StaticPool) Config() Config {
	return p.cfg
}

// Workers returns worker list associated with the pool.
func (p *StaticPool) Workers() (workers []*Worker) {
	p.muw.RLock()
	defer p.muw.RUnlock()

	for _, w := range p.workers {
		workers = append(workers, w)
	}

	return workers
}

// Remove forces pool to remove specific worker.
func (p *StaticPool) Remove(w *Worker, err error) bool {
	if w.State().Value() != StateReady && w.State().Value() != StateWorking {
		// unable to remove inactive worker
		return false
	}

	if _, ok := p.remove.Load(w); ok {
		return false
	}

	p.remove.Store(w, err)
	return true
}

// Exec one task with given payload and context, returns result or error.
func (p *StaticPool) Exec(rqs *Payload) (rsp *Payload, err error) {
	p.tmu.Lock()
	p.tasks.Add(1)
	p.tmu.Unlock()

	defer p.tasks.Done()

	w, err := p.allocateWorker()
	if err != nil {
		return nil, errors.Wrap(err, "unable to allocate worker")
	}

	rsp, err = w.Exec(rqs)

	if err != nil {
		// soft job errors are allowed
		if _, jobError := err.(JobError); jobError {
			p.release(w)
			return nil, err
		}

		p.discardWorker(w, err)
		return nil, err
	}

	// worker want's to be terminated
	if rsp.Body == nil && rsp.Context != nil && string(rsp.Context) == StopRequest {
		p.discardWorker(w, err)
		return p.Exec(rqs)
	}

	p.release(w)
	return rsp, nil
}

// Destroy all underlying workers (but let them to complete the task).
func (p *StaticPool) Destroy() {
	atomic.AddInt32(&p.inDestroy, 1)

	p.tmu.Lock()
	p.tasks.Wait()
	close(p.destroy)
	p.tmu.Unlock()

	var wg sync.WaitGroup
	for _, w := range p.Workers() {
		wg.Add(1)
		w.markInvalid()
		go func(w *Worker) {
			defer wg.Done()
			p.destroyWorker(w, nil)
		}(w)
	}

	wg.Wait()
}

// finds free worker in a given time interval. Skips dead workers.
func (p *StaticPool) allocateWorker() (w *Worker, err error) {
	for i := atomic.LoadInt64(&p.numDead); i >= 0; i++ {
		// this loop is required to skip issues with dead workers still being in a ring
		// (we know how many workers).
		select {
		case w = <-p.free:
			if w.State().Value() != StateReady {
				// found expected dead worker
				atomic.AddInt64(&p.numDead, ^int64(0))
				continue
			}

			if err, remove := p.remove.Load(w); remove {
				p.discardWorker(w, err)

				// get next worker
				i++
				continue
			}

			return w, nil
		case <-p.destroy:
			return nil, fmt.Errorf("pool has been stopped")
		default:
			// enable timeout handler
		}

		timeout := time.NewTimer(p.cfg.AllocateTimeout)
		select {
		case <-timeout.C:
			return nil, fmt.Errorf("worker timeout (%s)", p.cfg.AllocateTimeout)
		case w = <-p.free:
			timeout.Stop()

			if w.State().Value() != StateReady {
				atomic.AddInt64(&p.numDead, ^int64(0))
				continue
			}

			if err, remove := p.remove.Load(w); remove {
				p.discardWorker(w, err)

				// get next worker
				i++
				continue
			}

			return w, nil
		case <-p.destroy:
			timeout.Stop()

			return nil, fmt.Errorf("pool has been stopped")
		}
	}

	return nil, fmt.Errorf("all workers are dead (%v)", p.cfg.NumWorkers)
}

// release releases or replaces the worker.
func (p *StaticPool) release(w *Worker) {
	if p.cfg.MaxJobs != 0 && w.State().NumExecs() >= p.cfg.MaxJobs {
		p.discardWorker(w, p.cfg.MaxJobs)
		return
	}

	if err, remove := p.remove.Load(w); remove {
		p.discardWorker(w, err)
		return
	}

	p.free <- w
}

// creates new worker using associated factory. automatically
// adds worker to the worker list (background)
func (p *StaticPool) createWorker() (*Worker, error) {
	w, err := p.factory.SpawnWorker(p.cmd())
	if err != nil {
		return nil, err
	}

	p.mul.Lock()
	if p.lsn != nil {
		w.err.Listen(p.lsn)
	}
	p.mul.Unlock()

	p.throw(EventWorkerConstruct, w)

	p.muw.Lock()
	p.workers = append(p.workers, w)
	p.muw.Unlock()

	go p.watchWorker(w)
	return w, nil
}

// gentry remove worker
func (p *StaticPool) discardWorker(w *Worker, caused interface{}) {
	w.markInvalid()
	go p.destroyWorker(w, caused)
}

// destroyWorker destroys workers and removes it from the pool.
func (p *StaticPool) destroyWorker(w *Worker, caused interface{}) {
	go w.Stop()

	select {
	case <-w.waitDone:
		// worker is dead
		p.throw(EventWorkerDestruct, w)

	case <-time.NewTimer(p.cfg.DestroyTimeout).C:
		// failed to stop process in given time
		if err := w.Kill(); err != nil {
			p.throw(EventWorkerError, WorkerError{Worker: w, Caused: err})
		}

		p.throw(EventWorkerKill, w)
	}
}

// watchWorker watches worker state and replaces it if worker fails.
func (p *StaticPool) watchWorker(w *Worker) {
	err := w.Wait()
	p.throw(EventWorkerDead, w)

	// detaching
	p.muw.Lock()
	for i, wc := range p.workers {
		if wc == w {
			p.workers = append(p.workers[:i], p.workers[i+1:]...)
			p.remove.Delete(w)
			break
		}
	}
	p.muw.Unlock()

	// registering a dead worker
	atomic.AddInt64(&p.numDead, 1)

	// worker have died unexpectedly, pool should attempt to replace it with alive version safely
	if err != nil {
		p.throw(EventWorkerError, WorkerError{Worker: w, Caused: err})
	}

	if !p.destroyed() {
		nw, err := p.createWorker()
		if err == nil {
			p.free <- nw
			return
		}

		// possible situation when major error causes all PHP scripts to die (for example dead DB)
		if len(p.Workers()) == 0 {
			p.throw(EventPoolError, err)
		} else {
			p.throw(EventWorkerError, WorkerError{Worker: w, Caused: err})
		}
	}
}

func (p *StaticPool) destroyed() bool {
	return atomic.LoadInt32(&p.inDestroy) != 0
}

// throw invokes event handler if any.
func (p *StaticPool) throw(event int, ctx interface{}) {
	p.mul.Lock()
	if p.lsn != nil {
		p.lsn(event, ctx)
	}
	p.mul.Unlock()
}
