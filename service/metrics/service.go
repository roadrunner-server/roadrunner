package metrics

// todo: declare metric at runtime

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner/service/rpc"
	"golang.org/x/sys/cpu"
)

const (
	// ID declares public service name.
	ID = "metrics"
	// maxHeaderSize declares max header size for prometheus server
	maxHeaderSize = 1024 * 1024 * 100 // 104MB
)

// Service to manage application metrics using Prometheus.
type Service struct {
	cfg        *Config
	log        *logrus.Logger
	mu         sync.Mutex
	http       *http.Server
	collectors sync.Map
	registry   *prometheus.Registry
}

// Init service.
func (s *Service) Init(cfg *Config, r *rpc.Service, log *logrus.Logger) (bool, error) {
	s.cfg = cfg
	s.log = log
	s.registry = prometheus.NewRegistry()

	s.registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	s.registry.MustRegister(prometheus.NewGoCollector())

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
	return s.registry.Register(c)
}

// MustRegister registers new collector or fails with panic.
func (s *Service) MustRegister(c prometheus.Collector) {
	s.registry.MustRegister(c)
}

// Serve prometheus metrics service.
func (s *Service) Serve() error {
	// register application specific metrics
	collectors, err := s.cfg.getCollectors()
	if err != nil {
		return err
	}

	for name, collector := range collectors {
		if err := s.registry.Register(collector); err != nil {
			return err
		}

		s.collectors.Store(name, collector)
	}

	s.mu.Lock()

	var topCipherSuites []uint16
	var defaultCipherSuitesTLS13 []uint16

	hasGCMAsmAMD64 := cpu.X86.HasAES && cpu.X86.HasPCLMULQDQ
	hasGCMAsmARM64 := cpu.ARM64.HasAES && cpu.ARM64.HasPMULL
	// Keep in sync with crypto/aes/cipher_s390x.go.
	hasGCMAsmS390X := cpu.S390X.HasAES && cpu.S390X.HasAESCBC && cpu.S390X.HasAESCTR && (cpu.S390X.HasGHASH || cpu.S390X.HasAESGCM)

	hasGCMAsm := hasGCMAsmAMD64 || hasGCMAsmARM64 || hasGCMAsmS390X

	if hasGCMAsm {
		// If AES-GCM hardware is provided then prioritise AES-GCM
		// cipher suites.
		topCipherSuites = []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		}
		defaultCipherSuitesTLS13 = []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
		}
	} else {
		// Without AES-GCM hardware, we put the ChaCha20-Poly1305
		// cipher suites first.
		topCipherSuites = []uint16{
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		}
		defaultCipherSuitesTLS13 = []uint16{
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
		}
	}

	DefaultCipherSuites := make([]uint16, 0, 22)
	DefaultCipherSuites = append(DefaultCipherSuites, topCipherSuites...)
	DefaultCipherSuites = append(DefaultCipherSuites, defaultCipherSuitesTLS13...)

	s.http = &http.Server{
		Addr:              s.cfg.Address,
		Handler:           promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{}),
		IdleTimeout:       time.Hour * 24,
		ReadTimeout:       time.Minute * 60,
		MaxHeaderBytes:    maxHeaderSize,
		ReadHeaderTimeout: time.Minute * 60,
		WriteTimeout:      time.Minute * 60,
		TLSConfig: &tls.Config{
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.CurveP384,
				tls.CurveP521,
				tls.X25519,
			},
			CipherSuites:             DefaultCipherSuites,
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
		},
	}
	s.mu.Unlock()

	err = s.http.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Stop prometheus metrics service.
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.http != nil {
		// gracefully stop server
		go func() {
			err := s.http.Shutdown(context.Background())
			if err != nil {
				// Function should be Stop() error
				s.log.Error(fmt.Errorf("error shutting down the metrics server: error %v", err))
			}
		}()
	}
}

// Collector returns application specific collector by name or nil if collector not found.
func (s *Service) Collector(name string) prometheus.Collector {
	collector, ok := s.collectors.Load(name)
	if !ok {
		return nil
	}

	return collector.(prometheus.Collector)
}
