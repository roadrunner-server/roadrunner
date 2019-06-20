package http

import (
	"context"
	"fmt"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service/env"
	"github.com/spiral/roadrunner/service/http/attributes"
	"github.com/spiral/roadrunner/service/rpc"
	"github.com/spiral/roadrunner/util"
	"golang.org/x/net/http2"
	"net/http"
	"net/http/fcgi"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

const (
	// ID contains default svc name.
	ID = "http"

	// EventInitSSL thrown at moment of https initialization. SSL server passed as context.
	EventInitSSL = 750
)

// http middleware type.
type middleware func(f http.HandlerFunc) http.HandlerFunc

// Service manages rr, http servers.
type Service struct {
	cfg        *Config
	env        env.Environment
	lsns       []func(event int, ctx interface{})
	mdwr       []middleware
	mu         sync.Mutex
	rr         *roadrunner.Server
	controller roadrunner.Controller
	handler    *Handler
	http       *http.Server
	https      *http.Server
	fcgi      *http.Server
}

// Attach attaches controller. Currently only one controller is supported.
func (s *Service) Attach(w roadrunner.Controller) {
	s.controller = w
}

// AddMiddleware adds new net/http mdwr.
func (s *Service) AddMiddleware(m middleware) {
	s.mdwr = append(s.mdwr, m)
}

// AddListener attaches server event controller.
func (s *Service) AddListener(l func(event int, ctx interface{})) {
	s.lsns = append(s.lsns, l)
}

// Init must return configure svc and return true if svc hasStatus enabled. Must return error in case of
// misconfiguration. Services must not be used without proper configuration pushed first.
func (s *Service) Init(cfg *Config, r *rpc.Service, e env.Environment) (bool, error) {
	s.cfg = cfg
	s.env = e

	if r != nil {
		if err := r.Register(ID, &rpcServer{s}); err != nil {
			return false, err
		}
	}

	if !cfg.EnableHTTP() && !cfg.EnableTLS() && !cfg.EnableFCGI() {
		return false, nil
	}

	return true, nil
}

// Serve serves the svc.
func (s *Service) Serve() error {
	s.mu.Lock()

	if s.env != nil {
		if err := s.env.Copy(s.cfg.Workers); err != nil {
			return nil
		}
	}

	s.cfg.Workers.SetEnv("RR_HTTP", "true")

	s.rr = roadrunner.NewServer(s.cfg.Workers)
	s.rr.Listen(s.throw)

	if s.controller != nil {
		s.rr.Attach(s.controller)
	}

	if s.cfg.EnableMiddlewares() {
		s.initMiddlewares()
	}

	s.handler = &Handler{cfg: s.cfg, rr: s.rr}
	s.handler.Listen(s.throw)

	if s.cfg.EnableHTTP() {
		s.http = &http.Server{Addr: s.cfg.Address, Handler: s}
	}

	if s.cfg.EnableTLS() {
		s.https = s.initSSL()

		if s.cfg.EnableHTTP2() {
			if err := s.InitHTTP2(); err != nil {
				return err
			}
		}
	}

	if s.cfg.EnableFCGI() {
		s.fcgi = &http.Server{Handler: s}
	}

	s.mu.Unlock()

	if err := s.rr.Start(); err != nil {
		return err
	}
	defer s.rr.Stop()

	err := make(chan error, 3)

	if s.http != nil {
		go func() {
			err <- s.http.ListenAndServe()
		}()
	}

	if s.https != nil {
		go func() {
			err <- s.https.ListenAndServeTLS(s.cfg.SSL.Cert, s.cfg.SSL.Key)
		}()
	}

	if s.fcgi != nil {
		go func() {
			err <- s.ListenAndServeFCGI()
		}()
	}

	return <-err
}

// Stop stops the http.
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.fcgi != nil {
		go s.fcgi.Shutdown(context.Background())
	}

	if s.https != nil {
		go s.https.Shutdown(context.Background())
	}

	if s.http != nil {
		go s.http.Shutdown(context.Background())
	}
}

// Server returns associated rr server (if any).
func (s *Service) Server() *roadrunner.Server {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.rr
}

func (s *Service) ListenAndServeFCGI() error {
	l, err := util.CreateListener(s.cfg.FCGI.Address);
	if err != nil {
		return err
	}

	err = fcgi.Serve(l, s.fcgi.Handler)
	if err != nil {
		return err
	}

	return nil
}

// ServeHTTP handles connection using set of middleware and rr PSR-7 server.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.https != nil && r.TLS == nil && s.cfg.SSL.Redirect {
		target := &url.URL{
			Scheme:   "https",
			Host:     s.tlsAddr(r.Host, false),
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
		}

		http.Redirect(w, r, target.String(), http.StatusTemporaryRedirect)
		return
	}

	r = attributes.Init(r)

	// chaining middleware
	f := s.handler.ServeHTTP
	for _, m := range s.mdwr {
		f = m(f)
	}
	f(w, r)
}

// Init https server.
func (s *Service) initSSL() *http.Server {
	server := &http.Server{Addr: s.tlsAddr(s.cfg.Address, true), Handler: s}
	s.throw(EventInitSSL, server)

	return server
}

func (s *Service) InitHTTP2() error {
	return http2.ConfigureServer(s.https, &http2.Server{
		MaxConcurrentStreams: s.cfg.HTTP2.MaxConcurrentStreams,
	})
}

// throw handles service, server and pool events.
func (s *Service) throw(event int, ctx interface{}) {
	for _, l := range s.lsns {
		l(event, ctx)
	}

	if event == roadrunner.EventServerFailure {
		// underlying rr server is dead
		s.Stop()
	}
}

// tlsAddr replaces listen or host port with port configured by SSL config.
func (s *Service) tlsAddr(host string, forcePort bool) string {
	// remove current forcePort first
	host = strings.Split(host, ":")[0]

	if forcePort || s.cfg.SSL.Port != 443 {
		host = fmt.Sprintf("%s:%v", host, s.cfg.SSL.Port)
	}

	return host
}

func (s *Service) headersMiddleware(f http.HandlerFunc) http.HandlerFunc {
	// Define the http.HandlerFunc
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.Middlewares.Headers.CustomRequestHeaders != nil {
			for k, v := range s.cfg.Middlewares.Headers.CustomRequestHeaders {
				r.Header.Add(k, v)
			}
		}

		if s.cfg.Middlewares.Headers.CustomResponseHeaders != nil {
			for k, v := range s.cfg.Middlewares.Headers.CustomResponseHeaders {
				w.Header().Set(k, v)
			}
		}

		f(w, r)
	}
}

func handlePreflight(w http.ResponseWriter, r *http.Request, options *CORSMiddlewareConfig)  {
	headers := w.Header()

	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")

	if options.AllowedOrigin != "" {
		headers.Set("Access-Control-Allow-Origin", options.AllowedOrigin)
	}

	if options.AllowedHeaders != "" {
		headers.Set("Access-Control-Allow-Headers", options.AllowedHeaders)
	}

	if options.AllowedMethods != "" {
		headers.Set("Access-Control-Allow-Methods", options.AllowedMethods)
	}

	if options.AllowCredentials != nil {
		headers.Set("Access-Control-Allow-Credentials", strconv.FormatBool(*options.AllowCredentials))
	}

	if options.MaxAge > 0 {
		headers.Set("Access-Control-Max-Age", strconv.Itoa(options.MaxAge))
	}
}

func (s *Service) corsMiddleware(f http.HandlerFunc) http.HandlerFunc {
	// Define the http.HandlerFunc
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			handlePreflight(w, r, s.cfg.Middlewares.CORS)
			w.WriteHeader(http.StatusOK);
		} else {
			f(w, r)
		}
	}
}

func (s *Service) initMiddlewares() error {
	if s.cfg.Middlewares.EnableHeaders() {
		s.AddMiddleware(s.headersMiddleware)
	}

	if s.cfg.Middlewares.EnableCORS() {
		s.AddMiddleware(s.corsMiddleware)
	}

	return nil
}
