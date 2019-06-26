package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// ID declares public service name.
const ID = "metrics"

// Service to manage application metrics using Prometheus.
type Service struct {
	cfg  *Config
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
	s.http = &http.Server{Addr: s.cfg.Address, Handler: promhttp.Handler()}

	return s.http.ListenAndServe()
}

// Stop prometheus metrics service.
func (s *Service) Stop() {
	// gracefully stop server
	go s.http.Shutdown(context.Background())
}
