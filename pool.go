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

const (
	// EventCreated thrown when new worker is spawned.
	EventCreated = iota

	// EventDestruct thrown before worker destruction.
	EventDestruct

	// EventError thrown any worker related even happen (error passed as context)
	EventError
)

// Pool controls worker creation, destruction and task routing.
type Pool struct {
	// Observer is optional callback to handle worker create/destruct/error events.
	Observer func(event int, w *Worker, ctx interface{})

	// pool behaviour
	cfg Config

	// worker command creator
	cmd func() *exec.Cmd

	// creates and connects to workers
	factory Factory

	// active task executions
	tasks sync.WaitGroup

	// workers circular allocation buffer
	free chan *Worker

	// protects state of worker list, does not affect allocation
	muw sync.RWMutex

	// all registered workers
	workers []*Worker
}

// NewPool creates new worker pool and task multiplexer. Pool will initiate with one worker.
func NewPool(cmd func() *exec.Cmd, factory Factory, cfg Config) (*Pool, error) {
	if err := cfg.Valid(); err != nil {
		return nil, errors.Wrap(err, "config error")
	}

	p := &Pool{
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

// Config returns associated pool configuration. Immutable.
func (p *Pool) Config() Config {
	return p.cfg
}

// Workers returns worker list associated with the pool.
func (p *Pool) Workers() (workers []*Worker) {
	p.muw.RLock()
	defer p.muw.RUnlock()

	for _, w := range p.workers {
		workers = append(workers, w)
	}

	return workers
}

// Exec one task with given payload and context, returns result or error.
func (p *Pool) Exec(rqs *Payload) (rsp *Payload, err error) {
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
			p.free <- w
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

	if p.cfg.MaxExecutions != 0 && w.State().NumExecs() >= p.cfg.MaxExecutions {
		go p.replaceWorker(w, p.cfg.MaxExecutions)
	} else {
		p.free <- w
	}

	return rsp, nil
}

// Destroy all underlying workers (but let them to complete the task).
func (p *Pool) Destroy() {
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
func (p *Pool) allocateWorker() (w *Worker, err error) {
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

// replaces dead or expired worker with new instance
func (p *Pool) replaceWorker(w *Worker, caused interface{}) {
	go p.destroyWorker(w)

	if nw, err := p.createWorker(); err != nil {
		p.throw(EventError, w, err)
	} else {
		p.free <- nw
	}
}

// destroy and remove worker from the pool.
func (p *Pool) destroyWorker(w *Worker) {
	p.throw(EventDestruct, w, nil)

	// detaching
	p.muw.Lock()
	for i, wc := range p.workers {
		if wc == w {
			p.workers = p.workers[:i+1]
			break
		}
	}
	p.muw.Unlock()

	go w.Stop()

	select {
	case <-w.waitDone:
		// worker is dead
	case <-time.NewTimer(p.cfg.DestroyTimeout).C:
		// failed to stop process
		if err := w.Kill(); err != nil {
			p.throw(EventError, w, err)
		}
	}
}

// creates new worker using associated factory. automatically
// adds worker to the worker list (background)
func (p *Pool) createWorker() (*Worker, error) {
	w, err := p.factory.SpawnWorker(p.cmd())
	if err != nil {
		return nil, err
	}

	p.throw(EventCreated, w, nil)

	go func(w *Worker) {
		if err := w.Wait(); err != nil {
			p.throw(EventError, w, err)
		}
	}(w)

	p.muw.Lock()
	defer p.muw.Unlock()

	p.workers = append(p.workers, w)

	return w, nil
}

// throw invokes event handler if any.
func (p *Pool) throw(event int, w *Worker, ctx interface{}) {
	if p.Observer != nil {
		p.Observer(event, w, ctx)
	}
}
