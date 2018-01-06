package roadrunner

import (
	"fmt"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// ContextTerminate must be sent by worker in control payload if worker want to die.
	ContextTerminate = "TERMINATE"
)

// Pool controls worker creation, destruction and task routing.
type Pool struct {
	cfg        Config           // pool behaviour
	cmd        func() *exec.Cmd // worker command creator
	factory    Factory          // creates and connects to workers
	numWorkers uint64           // current number of tasks workers
	tasks      sync.WaitGroup   // counts all tasks executions
	mua        sync.Mutex       // protects worker allocation
	muw        sync.RWMutex     // protects st of worker list
	workers    []*Worker        // all registered workers
	free       chan *Worker     // freed workers
}

// NewPool creates new worker pool and task multiplexer. Pool will initiate with one worker.
func NewPool(cmd func() *exec.Cmd, factory Factory, cfg Config) (*Pool, error) {
	p := &Pool{
		cfg:     cfg,
		cmd:     cmd,
		factory: factory,
		workers: make([]*Worker, 0, cfg.MaxWorkers),
		free:    make(chan *Worker, cfg.MaxWorkers),
	}

	// to test if worker ready
	w, err := p.createWorker()
	if err != nil {
		return nil, err
	}

	p.free <- w
	return p, nil
}

// Exec one task with given payload and context, returns result and context or error. Must not be used once pool is
// being destroyed.
func (p *Pool) Exec(payload []byte, ctx interface{}) (resp []byte, rCtx []byte, err error) {
	p.tasks.Add(1)
	defer p.tasks.Done()

	w, err := p.allocateWorker()
	if err != nil {
		return nil, nil, err
	}

	if resp, rCtx, err = w.Exec(payload, ctx); err != nil {
		if !p.cfg.DestroyOnError {
			if err, jobError := err.(JobError); jobError {
				p.free <- w
				return nil, nil, err
			}
		}

		// worker level error
		p.destroyWorker(w)

		return nil, nil, err
	}

	// controlled destruction
	if len(resp) == 0 && string(rCtx) == ContextTerminate {
		p.destroyWorker(w)
		go func() {
			//immediate refill
			if w, err := p.createWorker(); err != nil {
				p.free <- w
			}
		}()

		return p.Execute(payload, ctx)
	}

	if p.cfg.MaxExecutions != 0 && atomic.LoadUint64(&w.numExecs) > p.cfg.MaxExecutions {
		p.destroyWorker(w)
	} else {
		p.free <- w
	}

	return resp, rCtx, nil
}

// Config returns associated pool configuration.
func (p *Pool) Config() Config {
	return p.cfg
}

// Workers returns workers associated with the pool.
func (p *Pool) Workers() (workers []*Worker) {
	p.muw.RLock()
	defer p.muw.RUnlock()

	for _, w := range p.workers {
		workers = append(workers, w)
	}

	return workers
}

// Close all underlying workers (but let them to complete the task).
func (p *Pool) Close() {
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
func (p *Pool) allocateWorker() (*Worker, error) {
	p.mua.Lock()
	defer p.mua.Unlock()

	select {
	case w := <-p.free:
		// we already have free worker
		return w, nil
	default:
		if p.numWorkers < p.cfg.MaxWorkers {
			return p.createWorker()
		}

		timeout := time.NewTimer(p.cfg.AllocateTimeout)
		select {
		case <-timeout.C:
			return nil, fmt.Errorf("unable to allocate worker, timeout (%st)", p.cfg.AllocateTimeout)
		case w := <-p.free:
			timeout.Stop()
			return w, nil
		}
	}
}

// destroy and remove worker from the pool.
func (p *Pool) destroyWorker(w *Worker) {
	atomic.AddUint64(&p.numWorkers, ^uint64(0))

	go func() {
		w.Stop()

		p.muw.Lock()
		defer p.muw.Unlock()

		for i, wc := range p.workers {
			if wc == w {
				p.workers = p.workers[:i+1]
				break
			}
		}
	}()
}

// creates new worker (must be called in a locked st).
func (p *Pool) createWorker() (*Worker, error) {
	w, err := p.factory.NewWorker(p.cmd())
	if err != nil {
		return nil, err
	}

	atomic.AddUint64(&p.numWorkers, 1)

	go func() {
		p.muw.Lock()
		defer p.muw.Unlock()
		p.workers = append(p.workers, w)
	}()

	return w, nil
}
