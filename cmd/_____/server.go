package roadrunner

import (
	"os/exec"
	"sync"
)

const (
	// EventNewPool triggered when server creates new pool.
	EventNewPool = 60

	// EventDestroyPool triggered when server destroys existed pool.
	EventDestroyPool = 61
)

// Service manages pool creation and swapping.
type Server struct {
	// configures server, pool, cmd creation and factory.
	scfg *ServerConfig

	// worker command creator
	cmd func() *exec.Cmd

	// observes pool events (can be attached to multiple pools at the same time)
	observer func(event int, ctx interface{})

	// creates and connects to workers
	factory Factory

	// protects pool while the switch
	mu sync.Mutex
}

// todo: do assignment

// Reconfigure configures underlying pool and destroys it's previous version if any.
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
