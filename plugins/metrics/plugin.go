package metrics

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"golang.org/x/sys/cpu"
)

const (
	// PluginName declares plugin name.
	PluginName = "metrics"
	// maxHeaderSize declares max header size for prometheus server
	maxHeaderSize = 1024 * 1024 * 100 // 104MB
)

// Plugin to manage application metrics using Prometheus.
type Plugin struct {
	cfg        *Config
	log        logger.Logger
	mu         sync.Mutex // all receivers are pointers
	http       *http.Server
	collectors sync.Map // all receivers are pointers
	registry   *prometheus.Registry
}

// Init service.
func (m *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("metrics_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &m.cfg)
	if err != nil {
		return errors.E(op, errors.Disabled, err)
	}

	m.cfg.InitDefaults()

	m.log = log
	m.registry = prometheus.NewRegistry()

	// Default
	err = m.registry.Register(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	if err != nil {
		return errors.E(op, err)
	}

	// Default
	err = m.registry.Register(prometheus.NewGoCollector())
	if err != nil {
		return errors.E(op, err)
	}

	collectors, err := m.cfg.getCollectors()
	if err != nil {
		return errors.E(op, err)
	}

	// Register invocation will be later in the Serve method
	for k, v := range collectors {
		m.collectors.Store(k, v)
	}
	return nil
}

// Register new prometheus collector.
func (m *Plugin) Register(c prometheus.Collector) error {
	return m.registry.Register(c)
}

// Serve prometheus metrics service.
func (m *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	m.collectors.Range(func(key, value interface{}) bool {
		// key - name
		// value - prometheus.Collector
		c := value.(prometheus.Collector)
		if err := m.registry.Register(c); err != nil {
			errCh <- err
			return false
		}

		return true
	})

	var topCipherSuites []uint16
	var defaultCipherSuitesTLS13 []uint16

	hasGCMAsmAMD64 := cpu.X86.HasAES && cpu.X86.HasPCLMULQDQ
	hasGCMAsmARM64 := cpu.ARM64.HasAES && cpu.ARM64.HasPMULL
	// Keep in sync with crypto/aes/cipher_s390x.go.
	hasGCMAsmS390X := cpu.S390X.HasAES && cpu.S390X.HasAESCBC && cpu.S390X.HasAESCTR && (cpu.S390X.HasGHASH || cpu.S390X.HasAESGCM)

	hasGCMAsm := hasGCMAsmAMD64 || hasGCMAsmARM64 || hasGCMAsmS390X

	if hasGCMAsm {
		// If AES-GCM hardware is provided then prioritize AES-GCM
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

	go func() {
		err := m.http.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
	}()

	return errCh
}

// Stop prometheus metrics service.
func (m *Plugin) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.http != nil {
		// timeout is 10 seconds
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		err := m.http.Shutdown(ctx)
		if err != nil {
			// Function should be Stop() error
			m.log.Error("stop error", "error", errors.Errorf("error shutting down the metrics server: error %v", err))
		}
	}
	return nil
}

// Collects used to collect all plugins which implement metrics.StatProvider interface (and Named)
func (m *Plugin) Collects() []interface{} {
	return []interface{}{
		m.AddStatProvider,
	}
}

// AddStatProvider adds a metrics provider
func (m *Plugin) AddStatProvider(stat StatProvider) error {
	for _, c := range stat.MetricsCollector() {
		err := m.registry.Register(c)
		if err != nil {
			return err
		}
	}

	return nil
}

// Name returns user friendly plugin name
func (m *Plugin) Name() string {
	return PluginName
}

// RPC interface satisfaction
func (m *Plugin) RPC() interface{} {
	return &rpcServer{
		svc: m,
		log: m.log,
	}
}
