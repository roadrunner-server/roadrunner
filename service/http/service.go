package http

import (
	"context"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service/env"
	"github.com/spiral/roadrunner/service/http/attributes"
	"github.com/spiral/roadrunner/service/rpc"
	"net/http"
	"sync"
	"sync/atomic"
)

// ID contains default svc name.
const ID = "http"

// http middleware type.
type middleware func(f http.HandlerFunc) http.HandlerFunc

// Service manages rr, http servers.
type Service struct {
	cfg      *Config
	env      env.Provider
	lsns     []func(event int, ctx interface{})
	mdws     []middleware
	mu       sync.Mutex
	rr       *roadrunner.Server
	stopping int32
	srv      *Handler
	http     *http.Server
}

// AddMiddleware adds new net/http middleware.
func (s *Service) AddMiddleware(m middleware) {
	s.mdws = append(s.mdws, m)
}

// AddListener attaches server event watcher.
func (s *Service) AddListener(l func(event int, ctx interface{})) {
	s.lsns = append(s.lsns, l)
}

// Init must return configure svc and return true if svc hasStatus enabled. Must return error in case of
// misconfiguration. Services must not be used without proper configuration pushed first.
func (s *Service) Init(cfg *Config, r *rpc.Service, e env.Provider) (bool, error) {
	if !cfg.Enable {
		return false, nil
	}

	s.cfg = cfg
	s.env = e
	if r != nil {
		r.Register(ID, &rpcServer{s})
	}

	return true, nil
}

// Serve serves the svc.
func (s *Service) Serve() error {
	s.mu.Lock()

	if s.env != nil {
		for k, v := range s.env.GetEnv() {
			s.cfg.Workers.SetEnv(k, v)
		}
	}

	rr := roadrunner.NewServer(s.cfg.Workers)

	s.rr = rr
	s.srv = &Handler{cfg: s.cfg, rr: s.rr}
	s.http = &http.Server{Addr: s.cfg.Address}

	s.rr.Listen(s.listener)
	s.srv.Listen(s.listener)

	if len(s.mdws) == 0 {
		s.http.Handler = s.srv
	} else {
		s.http.Handler = s
	}
	s.mu.Unlock()

	if err := rr.Start(); err != nil {
		return err
	}
	defer s.rr.Stop()

	return s.http.ListenAndServe()
}

// Stop stops the svc.
func (s *Service) Stop() {
	if atomic.LoadInt32(&s.stopping) != 0 {
		// already stopping
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.http == nil {
		return
	}

	s.http.Shutdown(context.Background())
}

// middleware handles connection using set of mdws and rr PSR-7 server.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = attributes.Init(r)

	// chaining middlewares
	f := s.srv.ServeHTTP
	for _, m := range s.mdws {
		f = m(f)
	}
	f(w, r)
}

// listener handles service, server and pool events.
func (s *Service) listener(event int, ctx interface{}) {
	for _, l := range s.lsns {
		l(event, ctx)
	}

	if event == roadrunner.EventServerFailure {
		// underlying rr server is dead
		s.Stop()
	}
}
