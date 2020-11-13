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
	"github.com/spiral/endure"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/interfaces/log"
	"github.com/spiral/roadrunner/v2/interfaces/metrics"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"golang.org/x/sys/cpu"
)

const (
	// ID declares public service name.
	ServiceName = "metrics"
	// maxHeaderSize declares max header size for prometheus server
	maxHeaderSize = 1024 * 1024 * 100 // 104MB
)

type statsProvider struct {
	collector prometheus.Collector
	name      string
}

// Plugin to manage application metrics using Prometheus.
type Plugin struct {
	cfg        Config
	log        log.Logger
	mu         sync.Mutex // all receivers are pointers
	http       *http.Server
	collectors sync.Map //[]statsProvider
	registry   *prometheus.Registry
}

// Init service.
func (m *Plugin) Init(cfg config.Configurer, log log.Logger) error {
	const op = errors.Op("Metrics Init")
	err := cfg.UnmarshalKey(ServiceName, &m.cfg)
	if err != nil {
		return err
	}

	//m.cfg.InitDefaults()

	m.log = log
	m.registry = prometheus.NewRegistry()

	err = m.registry.Register(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	if err != nil {
		return errors.E(op, err)
	}
	err = m.registry.Register(prometheus.NewGoCollector())
	if err != nil {
		return errors.E(op, err)
	}

	//m.collectors = make([]statsProvider, 0, 2)

	//if r != nil {
	//	if err := r.Register(ID, &rpcServer{s}); err != nil {
	//		return false, err
	//	}
	//}

	return nil
}

// Enabled indicates that server is able to collect metrics.
//func (m *Plugin) Enabled() bool {
//	return m.cfg != nil
//}
//
// Register new prometheus collector.
func (m *Plugin) Register(c prometheus.Collector) error {
	return m.registry.Register(c)
}

// MustRegister registers new collector or fails with panic.
//func (m *Plugin) MustRegister(c prometheus.Collector) {
//	m.registry.MustRegister(c)
//}

// Serve prometheus metrics service.
func (m *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	// register application specific metrics
	//collectors, err := m.cfg.getCollectors()
	//if err != nil {
	//	return err
	//}

	m.collectors.Range(func(key, value interface{}) bool {
		// key - name
		// value - collector
		c := value.(statsProvider)
		if err := m.registry.Register(c.collector); err != nil {
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

func (m *Plugin) Collects() []interface{} {
	return []interface{}{
		m.AddStatProvider,
	}
}

// Collector returns application specific collector by name or nil if collector not found.
func (m *Plugin) AddStatProvider(name endure.Named, stat metrics.StatProvider) error {
	m.collectors.Store(name.Name(), statsProvider{
		collector: stat.MetricsCollector(),
		name:      name.Name(),
	})
	return nil
}

func (m *Plugin) Name() string {
	return ServiceName
}

func (m *Plugin) RPC() interface{} {
	return &rpcServer{svc: m}
}
