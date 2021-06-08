package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/pkg/process"
	"github.com/spiral/roadrunner/v2/pkg/worker"
	handler "github.com/spiral/roadrunner/v2/pkg/worker_handler"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/http/attributes"
	httpConfig "github.com/spiral/roadrunner/v2/plugins/http/config"
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

	// metrics
	workersExporter  *workersExporter
	requestsExporter *requestsExporter

	// servers
	http  *http.Server
	https *http.Server
	fcgi  *http.Server
}

// Init must return configure svc and return true if svc hasStatus enabled. Must return error in case of
// misconfiguration. Services must not be used without proper configuration pushed first.
func (p *Plugin) Init(cfg config.Configurer, rrLogger logger.Logger, server server.Server) error {
	const op = errors.Op("http_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &p.cfg)
	if err != nil {
		return errors.E(op, err)
	}

	err = p.cfg.InitDefaults()
	if err != nil {
		return errors.E(op, err)
	}

	// rr logger (via plugin)
	p.log = rrLogger
	// use time and date in UTC format
	p.stdLog = log.New(logger.NewStdAdapter(p.log), "http_plugin: ", log.Ldate|log.Ltime|log.LUTC)

	p.mdwr = make(map[string]Middleware)

	if !p.cfg.EnableHTTP() && !p.cfg.EnableTLS() && !p.cfg.EnableFCGI() {
		return errors.E(op, errors.Disabled)
	}

	// init if nil
	if p.cfg.Env == nil {
		p.cfg.Env = make(map[string]string)
	}

	// initialize workersExporter
	p.workersExporter = newWorkersExporter()
	// initialize requests exporter
	p.requestsExporter = newRequestsExporter()

	p.cfg.Env[RrMode] = "http"
	p.server = server

	return nil
}

func (p *Plugin) logCallback(event interface{}) {
	if ev, ok := event.(handler.ResponseEvent); ok {
		p.log.Debug(fmt.Sprintf("%d %s %s", ev.Response.Status, ev.Request.Method, ev.Request.URI),
			"remote", ev.Request.RemoteAddr,
			"elapsed", ev.Elapsed().String(),
		)
	}
}

// Serve serves the svc.
func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 2)
	// run whole process in the goroutine
	go func() {
		// protect http initialization
		p.Lock()
		p.serve(errCh)
		p.Unlock()
	}()

	return errCh
}

func (p *Plugin) serve(errCh chan error) {
	var err error
	const op = errors.Op("http_plugin_serve")
	p.pool, err = p.server.NewWorkerPool(context.Background(), pool.Config{
		Debug:           p.cfg.Pool.Debug,
		NumWorkers:      p.cfg.Pool.NumWorkers,
		MaxJobs:         p.cfg.Pool.MaxJobs,
		AllocateTimeout: p.cfg.Pool.AllocateTimeout,
		DestroyTimeout:  p.cfg.Pool.DestroyTimeout,
		Supervisor:      p.cfg.Pool.Supervisor,
	}, p.cfg.Env, p.logCallback)
	if err != nil {
		errCh <- errors.E(op, err)
		return
	}

	p.handler, err = handler.NewHandler(
		p.cfg.MaxRequestSize,
		p.cfg.InternalErrorCode,
		*p.cfg.Uploads,
		p.cfg.Cidrs,
		p.pool,
	)
	if err != nil {
		errCh <- errors.E(op, err)
		return
	}

	p.handler.AddListener(p.logCallback, p.metricsCallback)

	if p.cfg.EnableHTTP() {
		if p.cfg.EnableH2C() {
			p.http = &http.Server{
				Handler:  h2c.NewHandler(p, &http2.Server{}),
				ErrorLog: p.stdLog,
			}
		} else {
			p.http = &http.Server{
				Handler:  p,
				ErrorLog: p.stdLog,
			}
		}
	}

	if p.cfg.EnableTLS() {
		p.https = p.initSSL()
		if p.cfg.SSLConfig.RootCA != "" {
			err = p.appendRootCa()
			if err != nil {
				errCh <- errors.E(op, err)
				return
			}
		}

		// if HTTP2Config not nil
		if p.cfg.HTTP2Config != nil {
			if err := p.initHTTP2(); err != nil {
				errCh <- errors.E(op, err)
				return
			}
		}
	}

	if p.cfg.EnableFCGI() {
		p.fcgi = &http.Server{Handler: p, ErrorLog: p.stdLog}
	}

	// start http, https and fcgi servers if requested in the config
	go func() {
		p.serveHTTP(errCh)
	}()

	go func() {
		p.serveHTTPS(errCh)
	}()

	go func() {
		p.serveFCGI(errCh)
	}()
}

// Stop stops the http.
func (p *Plugin) Stop() error {
	p.Lock()
	defer p.Unlock()

	if p.fcgi != nil {
		err := p.fcgi.Shutdown(context.Background())
		if err != nil && err != http.ErrServerClosed {
			p.log.Error("fcgi shutdown", "error", err)
		}
	}

	if p.https != nil {
		err := p.https.Shutdown(context.Background())
		if err != nil && err != http.ErrServerClosed {
			p.log.Error("https shutdown", "error", err)
		}
	}

	if p.http != nil {
		err := p.http.Shutdown(context.Background())
		if err != nil && err != http.ErrServerClosed {
			p.log.Error("http shutdown", "error", err)
		}
	}

	// check for safety
	if p.pool != nil {
		p.pool.Destroy(context.Background())
	}

	return nil
}

// ServeHTTP handles connection using set of middleware and pool PSR-7 server.
func (p *Plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := r.Body.Close()
		if err != nil {
			p.log.Error("body close", "error", err)
		}
	}()
	if headerContainsUpgrade(r) {
		http.Error(w, "server does not support upgrade header", http.StatusInternalServerError)
		return
	}

	if p.https != nil && r.TLS == nil && p.cfg.SSLConfig.Redirect {
		p.redirect(w, r)
		return
	}

	if p.https != nil && r.TLS != nil {
		w.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	}

	r = attributes.Init(r)
	// protect the case, when user sendEvent Reset and we are replacing handler with pool
	p.RLock()
	p.handler.ServeHTTP(w, r)
	p.RUnlock()
}

// Workers returns slice with the process states for the workers
func (p *Plugin) Workers() []process.State {
	p.RLock()
	defer p.RUnlock()

	workers := p.workers()

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
func (p *Plugin) workers() []worker.BaseProcess {
	return p.pool.Workers()
}

// Name returns endure.Named interface implementation
func (p *Plugin) Name() string {
	return PluginName
}

// Reset destroys the old pool and replaces it with new one, waiting for old pool to die
func (p *Plugin) Reset() error {
	p.Lock()
	defer p.Unlock()
	const op = errors.Op("http_plugin_reset")
	p.log.Info("HTTP plugin got restart request. Restarting...")
	p.pool.Destroy(context.Background())
	p.pool = nil

	var err error
	p.pool, err = p.server.NewWorkerPool(context.Background(), pool.Config{
		Debug:           p.cfg.Pool.Debug,
		NumWorkers:      p.cfg.Pool.NumWorkers,
		MaxJobs:         p.cfg.Pool.MaxJobs,
		AllocateTimeout: p.cfg.Pool.AllocateTimeout,
		DestroyTimeout:  p.cfg.Pool.DestroyTimeout,
		Supervisor:      p.cfg.Pool.Supervisor,
	}, p.cfg.Env, p.logCallback)
	if err != nil {
		return errors.E(op, err)
	}

	p.log.Info("HTTP workers Pool successfully restarted")

	p.handler, err = handler.NewHandler(
		p.cfg.MaxRequestSize,
		p.cfg.InternalErrorCode,
		*p.cfg.Uploads,
		p.cfg.Cidrs,
		p.pool,
	)

	if err != nil {
		return errors.E(op, err)
	}

	p.log.Info("HTTP handler listeners successfully re-added")
	p.handler.AddListener(p.logCallback, p.metricsCallback)

	p.log.Info("HTTP plugin successfully restarted")
	return nil
}

// Collects collecting http middlewares
func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.AddMiddleware,
	}
}

// AddMiddleware is base requirement for the middleware (name and Middleware)
func (p *Plugin) AddMiddleware(name endure.Named, m Middleware) {
	p.mdwr[name.Name()] = m
}

// Status return status of the particular plugin
func (p *Plugin) Status() status.Status {
	p.RLock()
	defer p.RUnlock()

	workers := p.workers()
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
func (p *Plugin) Ready() status.Status {
	p.RLock()
	defer p.RUnlock()

	workers := p.workers()
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
func (p *Plugin) Available() {}
