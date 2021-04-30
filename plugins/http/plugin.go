package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/pkg/process"
	"github.com/spiral/roadrunner/v2/pkg/worker"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/http/attributes"
	httpConfig "github.com/spiral/roadrunner/v2/plugins/http/config"
	"github.com/spiral/roadrunner/v2/plugins/http/static"
	handler "github.com/spiral/roadrunner/v2/plugins/http/worker_handler"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/spiral/roadrunner/v2/plugins/status"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	// PluginName declares plugin name.
	PluginName = "http"

	// RrMode RR_HTTP env variable key (internal) if the HTTP presents
	RrMode = "RR_MODE"

	HTTPSScheme = "https"
)

// Middleware interface
type Middleware interface {
	Middleware(f http.Handler) http.Handler
}

type middleware map[string]Middleware

// Plugin manages pool, http servers. The main http plugin structure
type Plugin struct {
	sync.RWMutex

	// plugins
	server server.Server
	log    logger.Logger
	// stdlog passed to the http/https/fcgi servers to log their internal messages
	stdLog *log.Logger

	// http configuration
	cfg *httpConfig.HTTP `mapstructure:"http"`

	// middlewares to chain
	mdwr middleware

	// Pool which attached to all servers
	pool pool.Pool

	// servers RR handler
	handler *handler.Handler

	// servers
	http  *http.Server
	https *http.Server
	fcgi  *http.Server
}

// Init must return configure svc and return true if svc hasStatus enabled. Must return error in case of
// misconfiguration. Services must not be used without proper configuration pushed first.
func (s *Plugin) Init(cfg config.Configurer, rrLogger logger.Logger, server server.Server) error {
	const op = errors.Op("http_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &s.cfg)
	if err != nil {
		return errors.E(op, err)
	}

	err = s.cfg.InitDefaults()
	if err != nil {
		return errors.E(op, err)
	}

	// rr logger (via plugin)
	s.log = rrLogger
	// use time and date in UTC format
	s.stdLog = log.New(logger.NewStdAdapter(s.log), "http_plugin: ", log.Ldate|log.Ltime|log.LUTC)

	s.mdwr = make(map[string]Middleware)

	if !s.cfg.EnableHTTP() && !s.cfg.EnableTLS() && !s.cfg.EnableFCGI() {
		return errors.E(op, errors.Disabled)
	}

	// init if nil
	if s.cfg.Env == nil {
		s.cfg.Env = make(map[string]string)
	}

	s.cfg.Env[RrMode] = "http"
	s.server = server

	return nil
}

func (s *Plugin) logCallback(event interface{}) {
	if ev, ok := event.(handler.ResponseEvent); ok {
		s.log.Debug(fmt.Sprintf("%d %s %s", ev.Response.Status, ev.Request.Method, ev.Request.URI),
			"remote", ev.Request.RemoteAddr,
			"elapsed", ev.Elapsed().String(),
		)
	}
}

// Serve serves the svc.
func (s *Plugin) Serve() chan error {
	errCh := make(chan error, 2)
	// run whole process in the goroutine
	go func() {
		// protect http initialization
		s.Lock()
		s.serve(errCh)
		s.Unlock()
	}()

	return errCh
}

func (s *Plugin) serve(errCh chan error) { //nolint:gocognit
	var err error
	const op = errors.Op("http_plugin_serve")
	s.pool, err = s.server.NewWorkerPool(context.Background(), pool.Config{
		Debug:           s.cfg.Pool.Debug,
		NumWorkers:      s.cfg.Pool.NumWorkers,
		MaxJobs:         s.cfg.Pool.MaxJobs,
		AllocateTimeout: s.cfg.Pool.AllocateTimeout,
		DestroyTimeout:  s.cfg.Pool.DestroyTimeout,
		Supervisor:      s.cfg.Pool.Supervisor,
	}, s.cfg.Env, s.logCallback)
	if err != nil {
		errCh <- errors.E(op, err)
		return
	}

	s.handler, err = handler.NewHandler(
		s.cfg.MaxRequestSize,
		*s.cfg.Uploads,
		s.cfg.Cidrs,
		s.pool,
	)
	if err != nil {
		errCh <- errors.E(op, err)
		return
	}

	s.handler.AddListener(s.logCallback)

	// Create new HTTP Multiplexer
	mux := http.NewServeMux()

	// if we have static, handler here, create a fileserver
	if s.cfg.Static != nil {
		h := http.FileServer(static.FS(s.cfg.Static))
		// Static files handler
		mux.HandleFunc(s.cfg.Static.Pattern, func(w http.ResponseWriter, r *http.Request) {
			if s.cfg.Static.Request != nil {
				for k, v := range s.cfg.Static.Request {
					r.Header.Add(k, v)
				}
			}

			if s.cfg.Static.Response != nil {
				for k, v := range s.cfg.Static.Response {
					w.Header().Set(k, v)
				}
			}

			// calculate etag for the resource
			if s.cfg.Static.CalculateEtag {
				// do not allow paths like ../../resource
				// only specified folder and resources in it
				// https://lgtm.com/rules/1510366186013/
				if strings.Contains(r.URL.Path, "..") {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				f, errS := os.Open(filepath.Join(s.cfg.Static.Dir, r.URL.Path))
				if errS != nil {
					s.log.Warn("error opening file to calculate the Etag", "provided path", r.URL.Path)
				}

				// Set etag value to the ResponseWriter
				static.SetEtag(s.cfg.Static, f, w)
			}

			h.ServeHTTP(w, r)
		})
	}

	// handle main route
	mux.HandleFunc("/", s.ServeHTTP)

	if s.cfg.EnableHTTP() {
		if s.cfg.EnableH2C() {
			s.http = &http.Server{Handler: h2c.NewHandler(mux, &http2.Server{}), ErrorLog: s.stdLog}
		} else {
			s.http = &http.Server{Handler: mux, ErrorLog: s.stdLog}
		}
	}

	if s.cfg.EnableTLS() {
		s.https = s.initSSL()
		if s.cfg.SSLConfig.RootCA != "" {
			err = s.appendRootCa()
			if err != nil {
				errCh <- errors.E(op, err)
				return
			}
		}

		// if HTTP2Config not nil
		if s.cfg.HTTP2Config != nil {
			if err := s.initHTTP2(); err != nil {
				errCh <- errors.E(op, err)
				return
			}
		}
	}

	if s.cfg.EnableFCGI() {
		s.fcgi = &http.Server{Handler: mux, ErrorLog: s.stdLog}
	}

	// start http, https and fcgi servers if requested in the config
	go func() {
		s.serveHTTP(errCh)
	}()

	go func() {
		s.serveHTTPS(errCh)
	}()

	go func() {
		s.serveFCGI(errCh)
	}()
}

// Stop stops the http.
func (s *Plugin) Stop() error {
	s.Lock()
	defer s.Unlock()

	var err error
	if s.fcgi != nil {
		err = s.fcgi.Shutdown(context.Background())
		if err != nil && err != http.ErrServerClosed {
			s.log.Error("error shutting down the fcgi server", "error", err)
			// write error and try to stop other transport
			err = multierror.Append(err)
		}
	}

	if s.https != nil {
		err = s.https.Shutdown(context.Background())
		if err != nil && err != http.ErrServerClosed {
			s.log.Error("error shutting down the https server", "error", err)
			// write error and try to stop other transport
			err = multierror.Append(err)
		}
	}

	if s.http != nil {
		err = s.http.Shutdown(context.Background())
		if err != nil && err != http.ErrServerClosed {
			s.log.Error("error shutting down the http server", "error", err)
			// write error and try to stop other transport
			err = multierror.Append(err)
		}
	}

	// check for safety
	if s.pool != nil {
		s.pool.Destroy(context.Background())
	}

	return err
}

// ServeHTTP handles connection using set of middleware and pool PSR-7 server.
func (s *Plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if headerContainsUpgrade(r) {
		http.Error(w, "server does not support upgrade header", http.StatusInternalServerError)
		return
	}

	if s.https != nil && r.TLS == nil && s.cfg.SSLConfig.Redirect {
		s.redirect(w, r)
		return
	}

	if s.https != nil && r.TLS != nil {
		w.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	}

	r = attributes.Init(r)
	// protect the case, when user sendEvent Reset and we are replacing handler with pool
	s.RLock()
	s.handler.ServeHTTP(w, r)
	s.RUnlock()
}

// Workers returns slice with the process states for the workers
func (s *Plugin) Workers() []process.State {
	s.RLock()
	defer s.RUnlock()

	workers := s.workers()

	ps := make([]process.State, 0, len(workers))
	for i := 0; i < len(workers); i++ {
		state, err := process.WorkerProcessState(workers[i])
		if err != nil {
			return nil
		}
		ps = append(ps, state)
	}

	return ps
}

// internal
func (s *Plugin) workers() []worker.BaseProcess {
	return s.pool.Workers()
}

// Name returns endure.Named interface implementation
func (s *Plugin) Name() string {
	return PluginName
}

// Reset destroys the old pool and replaces it with new one, waiting for old pool to die
func (s *Plugin) Reset() error {
	s.Lock()
	defer s.Unlock()
	const op = errors.Op("http_plugin_reset")
	s.log.Info("HTTP plugin got restart request. Restarting...")
	s.pool.Destroy(context.Background())
	s.pool = nil

	var err error
	s.pool, err = s.server.NewWorkerPool(context.Background(), pool.Config{
		Debug:           s.cfg.Pool.Debug,
		NumWorkers:      s.cfg.Pool.NumWorkers,
		MaxJobs:         s.cfg.Pool.MaxJobs,
		AllocateTimeout: s.cfg.Pool.AllocateTimeout,
		DestroyTimeout:  s.cfg.Pool.DestroyTimeout,
		Supervisor:      s.cfg.Pool.Supervisor,
	}, s.cfg.Env, s.logCallback)
	if err != nil {
		return errors.E(op, err)
	}

	s.log.Info("HTTP workers Pool successfully restarted")

	s.handler, err = handler.NewHandler(
		s.cfg.MaxRequestSize,
		*s.cfg.Uploads,
		s.cfg.Cidrs,
		s.pool,
	)
	if err != nil {
		return errors.E(op, err)
	}

	s.log.Info("HTTP handler listeners successfully re-added")
	s.handler.AddListener(s.logCallback)

	s.log.Info("HTTP plugin successfully restarted")
	return nil
}

// Collects collecting http middlewares
func (s *Plugin) Collects() []interface{} {
	return []interface{}{
		s.AddMiddleware,
	}
}

// AddMiddleware is base requirement for the middleware (name and Middleware)
func (s *Plugin) AddMiddleware(name endure.Named, m Middleware) {
	s.mdwr[name.Name()] = m
}

// Status return status of the particular plugin
func (s *Plugin) Status() status.Status {
	s.RLock()
	defer s.RUnlock()

	workers := s.workers()
	for i := 0; i < len(workers); i++ {
		if workers[i].State().IsActive() {
			return status.Status{
				Code: http.StatusOK,
			}
		}
	}
	// if there are no workers, threat this as error
	return status.Status{
		Code: http.StatusServiceUnavailable,
	}
}

// Ready return readiness status of the particular plugin
func (s *Plugin) Ready() status.Status {
	s.RLock()
	defer s.RUnlock()

	workers := s.workers()
	for i := 0; i < len(workers); i++ {
		// If state of the worker is ready (at least 1)
		// we assume, that plugin's worker pool is ready
		if workers[i].State().Value() == worker.StateReady {
			return status.Status{
				Code: http.StatusOK,
			}
		}
	}
	// if there are no workers, threat this as no content error
	return status.Status{
		Code: http.StatusServiceUnavailable,
	}
}

// Available interface implementation
func (s *Plugin) Available() {}
