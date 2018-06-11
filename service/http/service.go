package http

import (
	"net/http"
	"github.com/spiral/roadrunner/service"
	"context"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service/rpc"
)

// Name contains default svc name.
const Name = "http"

type Middleware interface {
	// Handle must return true if request/response pair is handled withing the middleware.
	Handle(w http.ResponseWriter, r *http.Request) bool
}

// Service manages rr, http servers.
type Service struct {
	cfg        *Config
	listeners  []func(event int, ctx interface{})
	middleware []Middleware
	rr         *roadrunner.Server
	srv        *Server
	http       *http.Server
}

func (s *Service) AddMiddleware(m Middleware) {
	s.middleware = append(s.middleware, m)
}

// AddListener attaches server event watcher.
func (s *Service) AddListener(l func(event int, ctx interface{})) {
	s.listeners = append(s.listeners, l)
}

// Configure must return configure svc and return true if svc hasStatus enabled. Must return error in case of
// misconfiguration. Services must not be used without proper configuration pushed first.
func (s *Service) Configure(cfg service.Config, c service.Container) (bool, error) {
	config := &Config{}
	if err := cfg.Unmarshal(config); err != nil {
		return false, err
	}

	if !config.Enable {
		return false, nil
	}

	if err := config.Valid(); err != nil {
		return false, err
	}

	s.cfg = config

	// registering http RPC interface
	if r, ok := c.Get(rpc.Name); ok >= service.StatusConfigured {
		if h, ok := r.(*rpc.Service); ok {
			h.Register(Name, &rpcServer{s})
		}
	}

	return true, nil
}

// Serve serves the svc.
func (s *Service) Serve() error {
	rr := roadrunner.NewServer(s.cfg.Workers)

	s.rr = rr
	s.srv = &Server{cfg: s.cfg, rr: s.rr}
	s.http = &http.Server{Addr: s.cfg.Address}

	s.rr.Listen(s.listener)
	s.srv.Listen(s.listener)

	if len(s.middleware) == 0 {
		s.http.Handler = s.srv
	} else {
		s.http.Handler = s
	}

	if err := rr.Start(); err != nil {
		return err
	}
	defer s.rr.Stop()

	if err := s.http.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

// Stop stops the svc.
func (s *Service) Stop() {
	if s.http == nil {
		return
	}

	s.http.Shutdown(context.Background())
}

// Handle handles connection using set of middleware and rr PSR-7 server.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, m := range s.middleware {
		if m.Handle(w, r) {
			return
		}
	}

	s.srv.ServeHTTP(w, r)
}

func (s *Service) listener(event int, ctx interface{}) {
	// todo: DIE on server failure

	for _, l := range s.listeners {
		l(event, ctx)
	}
}
