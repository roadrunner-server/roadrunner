package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"sync"
)

// ID declares public service name.
const ID = "metrics"

// Service to manage application metrics using Prometheus.
type Service struct {
	cfg  *Config
	mu   sync.Mutex
	http *http.Server
}

// Init service.
func (s *Service) Init(cfg *Config) (bool, error) {
	s.cfg = cfg
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
func (s *Service) Serve() error {
	// register application specific metrics
	if err := s.cfg.registerMetrics(); err != nil {
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
