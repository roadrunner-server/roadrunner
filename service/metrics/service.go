package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spiral/roadrunner/service/rpc"
	"net/http"
	"sync"
)

// ID declares public service name.
const ID = "metrics"

// Service to manage application metrics using Prometheus.
type Service struct {
	cfg        *Config
	mu         sync.Mutex
	http       *http.Server
	collectors map[string]prometheus.Collector
}

// Init service.
func (s *Service) Init(cfg *Config, r *rpc.Service) (bool, error) {
	s.cfg = cfg

	if r != nil {
		if err := r.Register(ID, &rpcServer{s}); err != nil {
			return false, err
		}
	}

	return true, nil
}

// Enabled indicates that server is able to collect metrics.
func (s *Service) Enabled() bool {
	return s.cfg != nil
}

// Register new prometheus collector.
func (s *Service) Register(c prometheus.Collector) error {
	return prometheus.Register(c)
}

// MustRegister registers new collector or fails with panic.
func (s *Service) MustRegister(c prometheus.Collector) {
	if err := prometheus.Register(c); err != nil {
		panic(err)
	}
}

// Serve prometheus metrics service.
func (s *Service) Serve() (err error) {
	// register application specific metrics
	if s.collectors, err = s.cfg.initCollectors(); err != nil {
		return err
	}

	s.mu.Lock()
	s.http = &http.Server{Addr: s.cfg.Address, Handler: promhttp.Handler()}
	s.mu.Unlock()

	return s.http.ListenAndServe()
}

// Stop prometheus metrics service.
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.http != nil {
		// gracefully stop server
		go s.http.Shutdown(context.Background())
	}
}

// Collector returns application specific collector by name or nil if collector not found.
func (s *Service) Collector(name string) prometheus.Collector {
	collector, ok := s.collectors[name]
	if !ok {
		return nil
	}

	return collector
}
