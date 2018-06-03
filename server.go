package roadrunner

import (
	"sync"
	"os/exec"
	"fmt"
)

const (
	// EventNewPool triggered when server creates new pool.
	EventNewPool = 60

	// EventDestroyPool triggered when server destroys existed pool.
	EventDestroyPool = 61
)

// Service manages pool creation and swapping.
type Server struct {
	// observes pool events (can be attached to multiple pools at the same time)
	observer func(event int, ctx interface{})

	// worker command creator
	cmd func() *exec.Cmd

	// pool behaviour
	cfg Config

	// creates and connects to workers
	factory Factory

	// protects pool while the switch
	mu sync.Mutex

	// currently active pool instance
	pool Pool
}

// NewServer creates new router. Make sure to call configure before the usage.
func NewServer(cmd func() *exec.Cmd, factory Factory) *Server {
	return &Server{
		cmd:     cmd,
		factory: factory,
	}
}

// Configure configures underlying pool and destroys it's previous version if any.
func (r *Server) Configure(cfg Config) error {
	r.mu.Lock()
	previous := r.pool
	r.mu.Unlock()

	pool, err := NewPool(r.cmd, r.factory, cfg)
	if err != nil {
		return err
	}

	r.throw(EventNewPool, pool)

	r.mu.Lock()

	r.cfg, r.pool = cfg, pool
	r.pool.Observe(r.poolObserver)

	r.mu.Unlock()

	if previous != nil {
		go func(p Pool) {
			r.throw(EventDestroyPool, p)
			p.Destroy()
		}(previous)
	}

	return nil
}

// Reset resets the state of underlying pool and rebuilds all of it's workers.
func (r *Server) Reset() error {
	return r.Configure(r.cfg)
}

// Observe attaches event watcher to the router.
func (r *Server) Observe(o func(event int, ctx interface{})) {
	r.observer = o
}

// Pool returns active pool or error.
func (r *Server) Pool() (Pool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.pool
}

// Exec one task with given payload and context, returns result or error.
func (r *Server) Exec(rqs *Payload) (rsp *Payload, err error) {
	pool := r.Pool()
	if pool == nil {
		return nil, fmt.Errorf("no associared pool")
	}

	return pool.Exec(rqs)
}

// Workers returns worker list associated with the pool.
func (r *Server) Workers() (workers []*Worker) {
	pool := r.Pool()
	if pool == nil {
		return nil
	}

	return pool.Workers()
}

// Destroy all underlying pools and workers workers (but let them to complete the task).
func (r *Server) Destroy() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.pool == nil {
		return
	}

	go func(p Pool) {
		r.throw(EventDestroyPool, p)
		p.Destroy()
	}(r.pool)

	r.pool = nil
}

// throw invokes event handler if any.
func (r *Server) throw(event int, ctx interface{}) {
	if r.observer != nil {
		r.observer(event, ctx)
	}
}

// Observe pool events.
func (r *Server) poolObserver(event int, ctx interface{}) {
	// bypassing to user specified observer
	r.throw(event, ctx)

	if event == EventPoolError {
		// pool failure, rebuilding
		r.Reset()
	}
}
