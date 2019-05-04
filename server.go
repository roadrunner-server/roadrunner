package roadrunner

import (
	"fmt"
	"github.com/pkg/errors"
	"sync"
)

const (
	// EventServerStart triggered when server creates new pool.
	EventServerStart = iota + 200

	// EventServerStop triggered when server creates new pool.
	EventServerStop

	// EventServerFailure triggered when server is unable to replace dead pool.
	EventServerFailure

	// EventPoolConstruct triggered when server creates new pool.
	EventPoolConstruct

	// EventPoolDestruct triggered when server destroys existed pool.
	EventPoolDestruct
)

// Server manages pool creation and swapping.
type Server struct {
	// configures server, pool, cmd creation and factory.
	cfg *ServerConfig

	// protects pool while the re-configuration
	mu sync.Mutex

	// indicates that server was started
	started bool

	// creates and connects to workers
	factory Factory

	// associated pool controller
	controller Controller

	// currently active pool instance
	mup         sync.Mutex
	pool        Pool
	pController Controller

	// observes pool events (can be attached to multiple pools at the same time)
	mul sync.Mutex
	lsn func(event int, ctx interface{})
}

// NewServer creates new router. Make sure to call configure before the usage.
func NewServer(cfg *ServerConfig) *Server {
	return &Server{cfg: cfg}
}

// Listen attaches server event controller.
func (s *Server) Listen(l func(event int, ctx interface{})) {
	s.mul.Lock()
	defer s.mul.Unlock()

	s.lsn = l
}

// Attach attaches worker controller.
func (s *Server) Attach(c Controller) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.controller = c

	s.mul.Lock()
	if s.pController != nil && s.pool != nil {
		s.pController.Detach()
		s.pController = s.controller.Attach(s.pool)
	}
	s.mul.Unlock()
}

// Start underlying worker pool, configure factory and command provider.
func (s *Server) Start() (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.factory, err = s.cfg.makeFactory(); err != nil {
		return err
	}

	if s.pool, err = NewPool(s.cfg.makeCommand(), s.factory, *s.cfg.Pool); err != nil {
		return err
	}

	if s.controller != nil {
		s.pController = s.controller.Attach(s.pool)
	}

	s.pool.Listen(s.poolListener)
	s.started = true
	s.throw(EventServerStart, s)

	return nil
}

// Stop underlying worker pool and close the factory.
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return
	}

	s.throw(EventPoolDestruct, s.pool)

	if s.pController != nil {
		s.pController.Detach()
		s.pController = nil
	}

	s.pool.Destroy()
	s.factory.Close()

	s.factory = nil
	s.pool = nil
	s.started = false
	s.throw(EventServerStop, s)
}

// Exec one task with given payload and context, returns result or error.
func (s *Server) Exec(rqs *Payload) (rsp *Payload, err error) {
	pool := s.Pool()
	if pool == nil {
		return nil, fmt.Errorf("no associared pool")
	}

	return pool.Exec(rqs)
}

// Reconfigure re-configures underlying pool and destroys it's previous version if any. Reconfigure will ignore factory
// and relay settings.
func (s *Server) Reconfigure(cfg *ServerConfig) error {
	s.mup.Lock()
	defer s.mup.Unlock()

	s.mu.Lock()
	if !s.started {
		s.cfg = cfg
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	if s.cfg.Differs(cfg) {
		return errors.New("unable to reconfigure server (cmd and pool changes are allowed)")
	}

	s.mu.Lock()
	previous := s.pool
	pWatcher := s.pController
	s.mu.Unlock()

	pool, err := NewPool(cfg.makeCommand(), s.factory, *cfg.Pool)
	if err != nil {
		return err
	}

	pool.Listen(s.poolListener)

	s.mu.Lock()
	s.cfg.Pool, s.pool = cfg.Pool, pool

	if s.controller != nil {
		s.pController = s.controller.Attach(pool)
	}

	s.mu.Unlock()

	s.throw(EventPoolConstruct, pool)

	if previous != nil {
		go func(previous Pool, pWatcher Controller) {
			s.throw(EventPoolDestruct, previous)
			if pWatcher != nil {
				pWatcher.Detach()
			}

			previous.Destroy()
		}(previous, pWatcher)
	}

	return nil
}

// Reset resets the state of underlying pool and rebuilds all of it's workers.
func (s *Server) Reset() error {
	s.mu.Lock()
	cfg := s.cfg
	s.mu.Unlock()

	return s.Reconfigure(cfg)
}

// Workers returns worker list associated with the server pool.
func (s *Server) Workers() (workers []*Worker) {
	p := s.Pool()
	if p == nil {
		return nil
	}

	return p.Workers()
}

// Pool returns active pool or error.
func (s *Server) Pool() Pool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.pool
}

// Listen pool events.
func (s *Server) poolListener(event int, ctx interface{}) {
	if event == EventPoolError {
		// pool failure, rebuilding
		if err := s.Reset(); err != nil {
			s.mu.Lock()
			s.started = false
			s.pool = nil
			s.factory = nil
			s.mu.Unlock()

			// everything is dead, this is recoverable but heavy state
			s.throw(EventServerFailure, err)
		}
	}

	// bypassing to user specified lsn
	s.throw(event, ctx)
}

// throw invokes event handler if any.
func (s *Server) throw(event int, ctx interface{}) {
	s.mul.Lock()
	if s.lsn != nil {
		s.lsn(event, ctx)
	}
	s.mul.Unlock()
}
