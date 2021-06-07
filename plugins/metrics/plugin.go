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

	// prometheus Collectors
	statProviders []StatProvider
}

// Init service.
func (p *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("metrics_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &p.cfg)
	if err != nil {
		return errors.E(op, errors.Disabled, err)
	}

	p.cfg.InitDefaults()

	p.log = log
	p.registry = prometheus.NewRegistry()

	// Default
	err = p.registry.Register(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	if err != nil {
		return errors.E(op, err)
	}

	// Default
	err = p.registry.Register(prometheus.NewGoCollector())
	if err != nil {
		return errors.E(op, err)
	}

	collectors, err := p.cfg.getCollectors()
	if err != nil {
		return errors.E(op, err)
	}

	// Register invocation will be later in the Serve method
	for k, v := range collectors {
		p.collectors.Store(k, v)
	}

	p.statProviders = make([]StatProvider, 0, 2)

	return nil
}

// Register new prometheus collector.
func (p *Plugin) Register(c prometheus.Collector) error {
	return p.registry.Register(c)
}

// Serve prometheus metrics service.
func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	// register Collected stat providers
	for i := 0; i < len(p.statProviders); i++ {
		sp := p.statProviders[i]
		for _, c := range sp.MetricsCollector() {
			err := p.registry.Register(c)
			if err != nil {
				errCh <- err
				return errCh
			}
		}
	}

	p.collectors.Range(func(key, value interface{}) bool {
		// key - name
		// value - prometheus.Collector
		c := value.(prometheus.Collector)
		if err := p.registry.Register(c); err != nil {
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

	p.http = &http.Server{
		Addr:              p.cfg.Address,
		Handler:           promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{}),
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
		err := p.http.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
	}()

	return errCh
}

// Stop prometheus metrics service.
func (p *Plugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.http != nil {
		// timeout is 10 seconds
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		err := p.http.Shutdown(ctx)
		if err != nil {
			// Function should be Stop() error
			p.log.Error("stop error", "error", errors.Errorf("error shutting down the metrics server: error %v", err))
		}
	}
	return nil
}

// Collects used to collect all plugins which implement metrics.StatProvider interface (and Named)
func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.AddStatProvider,
	}
}

// AddStatProvider adds a metrics provider
func (p *Plugin) AddStatProvider(stat StatProvider) error {
	p.statProviders = append(p.statProviders, stat)

	return nil
}

// Name returns user friendly plugin name
func (p *Plugin) Name() string {
	return PluginName
}

// RPC interface satisfaction
func (p *Plugin) RPC() interface{} {
	return &rpcServer{
		svc: p,
		log: p.log,
	}
}

// Available interface implementation
func (p *Plugin) Available() {}
