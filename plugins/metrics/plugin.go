package metrics

// todo: declare metric at runtime

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/log"
	"github.com/spiral/roadrunner/v2/metrics"
	"golang.org/x/sys/cpu"
)

const (
	// ID declares public service name.
	ID = "metrics"
	// maxHeaderSize declares max header size for prometheus server
	maxHeaderSize = 1024 * 1024 * 100 // 104MB
)

// Plugin to manage application metrics using Prometheus.
type Plugin struct {
	cfg        Config
	log        log.Logger
	mu         sync.Mutex // all receivers are pointers
	http       *http.Server
	collectors []prometheus.Collector //sync.Map // all receivers are pointers
	registry   *prometheus.Registry
}

// Init service.
func (m *Plugin) Init(cfg Config, log log.Logger) (bool, error) {
	m.cfg = cfg
	m.log = log
	m.registry = prometheus.NewRegistry()

	m.registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	m.registry.MustRegister(prometheus.NewGoCollector())

	//if r != nil {
	//	if err := r.Register(ID, &rpcServer{s}); err != nil {
	//		return false, err
	//	}
	//}

	return true, nil
}

// Enabled indicates that server is able to collect metrics.
//func (m *Plugin) Enabled() bool {
//	return m.cfg != nil
//}
//
//// Register new prometheus collector.
//func (m *Plugin) Register(c prometheus.Collector) error {
//	return m.registry.Register(c)
//}

// MustRegister registers new collector or fails with panic.
func (m *Plugin) MustRegister(c prometheus.Collector) {
	m.registry.MustRegister(c)
}

// Serve prometheus metrics service.
func (m *Plugin) Serve() error {
	// register application specific metrics
	collectors, err := m.cfg.getCollectors()
	if err != nil {
		return err
	}

	for name, collector := range collectors {
		if err := m.registry.Register(collector); err != nil {
			return err
		}

		m.collectors.Store(name, collector)
	}

	m.mu.Lock()

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

	m.http = &http.Server{
		Addr:              m.cfg.Address,
		Handler:           promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}),
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
	m.mu.Unlock()

	err = m.http.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Stop prometheus metrics service.
func (m *Plugin) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.http != nil {
		// gracefully stop server
		go func() {
			err := m.http.Shutdown(context.Background())
			if err != nil {
				// Function should be Stop() error
				m.log.Error("stop error", "error", errors.Errorf("error shutting down the metrics server: error %v", err))
			}
		}()
	}
}

func (m *Plugin) Collects() []interface{} {
	return []interface{}{
		m.Register,
	}
}

// Collector returns application specific collector by name or nil if collector not found.
func (m *Plugin) Register(stat metrics.StatProvider) error {
	m.collectors = append(m.collectors, stat.MetricsCollector())
	return nil
}
