package roadrunner

import (
	"sync"
	"os/exec"
	"fmt"
)

const (
	// EventNewPool triggered when server creates new pool.
	EventServerStart = iota + 128

	// EventNewPool triggered when server creates new pool.
	EventServerStop

	// EventServerFailure triggered when server is unable to replace dead pool.
	EventServerFailure

	// EventReplaceFailure triggered when server can not replace pool while the re-configuration.
	EventReplaceFailure

	// EventNewPool triggered when server creates new pool.
	EventNewPool

	// EventDestroyPool triggered when server destroys existed pool.
	EventDestroyPool
)

// Service manages pool creation and swapping.
type Server struct {
	// configures server, pool, cmd creation and factory.
	cfg *ServerConfig

	// worker command creator
	cmd func() *exec.Cmd

	// observes pool events (can be attached to multiple pools at the same time)
	observer func(event int, ctx interface{})

	// protects pool while the re-configuration
	mu sync.Mutex

	// indicates that server was started
	started bool

	// creates and connects to workers
	factory Factory

	// currently active pool instance
	pool Pool
}

// NewServer creates new router. Make sure to call configure before the usage.
func NewServer(cmd func() *exec.Cmd, cfg *ServerConfig) *Server {
	return &Server{cmd: cmd, cfg: cfg}
}

// Observe attaches server event watcher.
func (srv *Server) Observe(o func(event int, ctx interface{})) {
	srv.observer = o
}

// Reconfigure re-configures underlying pool and destroys it's previous version if any. Reconfigure will ignore factory
// and relay settings.
func (srv *Server) Reconfigure(cfg *ServerConfig) error {
	srv.mu.Lock()
	if !srv.started {
		srv.cfg = cfg
		srv.mu.Unlock()
		return nil
	}
	srv.mu.Unlock()

	srv.mu.Lock()
	previous := srv.pool
	srv.mu.Unlock()

	pool, err := NewPool(srv.cmd, srv.factory, cfg.Pool)
	if err != nil {
		srv.throw(EventReplaceFailure, err)
		return err
	}
	srv.throw(EventNewPool, pool)

	srv.mu.Lock()
	srv.cfg.Pool, srv.pool = cfg.Pool, pool
	srv.pool.Observe(srv.poolObserver)
	srv.mu.Unlock()

	if previous != nil {
		go func(previous Pool) {
			srv.throw(EventDestroyPool, previous)
			previous.Destroy()
		}(previous)
	}

	return nil
}

// Reset resets the state of underlying pool and rebuilds all of it's workers.
func (srv *Server) Reset() error {
	return srv.Reconfigure(srv.cfg)
}

// Start underlying worker pool, configure factory and command provider.
func (srv *Server) Start() (err error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if srv.factory, err = srv.cfg.makeFactory(); err != nil {
		return err
	}

	if srv.pool, err = NewPool(srv.cmd, srv.factory, srv.cfg.Pool); err != nil {
		return err
	}

	srv.pool.Observe(srv.poolObserver)
	srv.started = true
	srv.throw(EventServerStart, srv)

	return nil
}

// Stop underlying worker pool and close the factory.
func (srv *Server) Stop() error {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if !srv.started {
		return nil
	}

	srv.throw(EventDestroyPool, srv.pool)
	srv.pool.Destroy()
	srv.factory.Close()

	srv.cmd = nil
	srv.factory = nil
	srv.pool = nil
	srv.started = false
	srv.throw(EventServerStop, srv)

	return nil
}

// Exec one task with given payload and context, returns result or error.
func (srv *Server) Exec(rqs *Payload) (rsp *Payload, err error) {
	pool := srv.Pool()
	if pool == nil {
		return nil, fmt.Errorf("no associared pool")
	}

	return pool.Exec(rqs)
}

// Workers returns worker list associated with the server pool.
func (srv *Server) Workers() (workers []*Worker) {
	p := srv.Pool()
	if p == nil {
		return nil
	}

	return p.Workers()
}

// Pool returns active pool or error.
func (srv *Server) Pool() Pool {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	return srv.pool
}

// Observe pool events.
func (srv *Server) poolObserver(event int, ctx interface{}) {
	// bypassing to user specified observer
	srv.throw(event, ctx)

	if event == EventPoolError {
		// pool failure, rebuilding
		if err := srv.Reset(); err != nil {
			srv.mu.Lock()
			defer srv.mu.Unlock()

			srv.started = false
			srv.pool = nil
			srv.factory = nil

			// everything is dead, this is recoverable but heavy state
			srv.throw(EventServerFailure, srv)
		}
	}
}

// throw invokes event handler if any.
func (srv *Server) throw(event int, ctx interface{}) {
	if srv.observer != nil {
		srv.observer(event, ctx)
	}
}
