package roadrunner

import (
	"fmt"
	"github.com/pkg/errors"
	"os/exec"
	"sync"
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

	// observer is optional callback to handle worker create/destruct/error events.
	observer func(event int, ctx interface{})

	// creates and connects to workers
	factory Factory

	// active task executions
	tasks sync.WaitGroup

	// workers circular allocation errBuffer
	free chan *Worker

	// protects state of worker list, does not affect allocation
	muw sync.RWMutex

	// all registered workers
	workers []*Worker
}

// NewPool creates new worker pool and task multiplexer. StaticPool will initiate with one worker.
func NewPool(cmd func() *exec.Cmd, factory Factory, cfg Config) (*StaticPool, error) {
	if err := cfg.Valid(); err != nil {
		return nil, errors.Wrap(err, "config error")
	}

	p := &StaticPool{
		cfg:     cfg,
		cmd:     cmd,
		factory: factory,
		workers: make([]*Worker, 0, cfg.NumWorkers),
		free:    make(chan *Worker, cfg.NumWorkers),
	}

	// constant number of workers simplify logic
	for i := uint64(0); i < p.cfg.NumWorkers; i++ {
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

// Observe attaches pool event watcher.
func (p *StaticPool) Observe(o func(event int, ctx interface{})) {
	p.observer = o
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

// Exec one task with given payload and context, returns result or error.
func (p *StaticPool) Exec(rqs *Payload) (rsp *Payload, err error) {
	p.tasks.Add(1)
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

		go p.replaceWorker(w, err)
		return nil, err
	}

	// worker want's to be terminated
	if rsp.Body == nil && rsp.Context != nil && string(rsp.Context) == StopRequest {
		go p.replaceWorker(w, err)
		return p.Exec(rqs)
	}

	p.release(w)
	return rsp, nil
}

// Destroy all underlying workers (but let them to complete the task).
func (p *StaticPool) Destroy() {
	p.tasks.Wait()

	var wg sync.WaitGroup
	for _, w := range p.Workers() {
		wg.Add(1)
		go func(w *Worker) {
			defer wg.Done()
			p.destroyWorker(w)
		}(w)
	}

	wg.Wait()
}

// finds free worker in a given time interval or creates new if allowed.
func (p *StaticPool) allocateWorker() (w *Worker, err error) {
	select {
	case w = <-p.free:
		return w, nil
	default:
		// enable timeout handler
	}

	timeout := time.NewTimer(p.cfg.AllocateTimeout)
	select {
	case <-timeout.C:
		return nil, fmt.Errorf("worker timeout (%s)", p.cfg.AllocateTimeout)
	case w := <-p.free:
		timeout.Stop()
		return w, nil
	}
}

// release releases or replaces the worker.
func (p *StaticPool) release(w *Worker) {
	if p.cfg.MaxExecutions != 0 && w.State().NumExecs() >= p.cfg.MaxExecutions {
		go p.replaceWorker(w, p.cfg.MaxExecutions)
		return
	}

	p.free <- w
}

// replaceWorker replaces dead or expired worker with new instance.
func (p *StaticPool) replaceWorker(w *Worker, caused interface{}) error {
	go p.destroyWorker(w)

	if nw, err := p.createWorker(); err != nil {
		p.throw(EventWorkerError, WorkerError{Worker: w, Caused: err})

		if len(p.Workers()) == 0 {
			// possible situation when major error causes all PHP scripts to die (for example dead DB)
			p.throw(EventPoolError, fmt.Errorf("all workers are dead"))
		}
		return err
	} else {
		p.free <- nw
	}

	return nil
}

// creates new worker using associated factory. automatically
// adds worker to the worker list (background)
func (p *StaticPool) createWorker() (*Worker, error) {
	w, err := p.factory.SpawnWorker(p.cmd())
	if err != nil {
		return nil, err
	}

	p.throw(EventWorkerCreate, w)

	go func(w *Worker) {
		err := w.Wait()

		// worker have died unexpectedly, pool should attempt to replace it with alive version safely
		if w.state.Value() == StateErrored {
			if err != nil {
				p.throw(EventWorkerError, WorkerError{Worker: w, Caused: err})
			}

			// attempting to replace worker
			if err := p.replaceWorker(w, err); err != nil {
				p.throw(EventPoolError, fmt.Errorf("unable to replace dead worker: %s", err))
			}
		}
	}(w)

	p.muw.Lock()
	defer p.muw.Unlock()

	p.workers = append(p.workers, w)

	return w, nil
}

// destroyWorker destroys workers and removes it from the pool.
func (p *StaticPool) destroyWorker(w *Worker) {
	// detaching
	p.muw.Lock()
	for i, wc := range p.workers {
		if wc == w {
			p.workers = append(p.workers[:i], p.workers[i+1:]...)
			break
		}
	}
	p.muw.Unlock()

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

// throw invokes event handler if any.
func (p *StaticPool) throw(event int, ctx interface{}) {
	if p.observer != nil {
		p.observer(event, ctx)
	}
}
