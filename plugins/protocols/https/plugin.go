package https

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/utils"
	"golang.org/x/sys/cpu"
)

// RootPluginName is the name of root plugin
const (
	RootPluginName string = "http"

	PluginName string = "ssl"

	HTTPSScheme = "https"
)

type Plugin struct {
	s      *Server
	server *http.Server
	cfg    *Config

	// this variable indicates if the ssl sub-plugin initialized in the configuration
	// if not - we just return nil server, but an error.
	initialized bool
}

// Server represents https server structure which include all needed to start serving
// It only needs a handler and ErrorLog endpoint to connect
// Handler should be ServeHTTP handler with worker handler
type Server struct {
	Server   *http.Server
	Listener net.Listener
	Redirect func(w http.ResponseWriter, r *http.Request)
	Key      string
	Cert     string
}

func (p *Plugin) Init(cfg config.Configurer) error {
	const op = errors.Op("https_plugin_init")
	// If there is no ssl activated, just return nil server
	// because we should not turn off the whole http server if sub-plugin not activated
	if !cfg.Has(fmt.Sprintf("%s.%s", RootPluginName, PluginName)) {
		// set not initialized state
		p.initialized = false
		return nil
	}

	// Unmarshal only section for the ssl sub-plugin
	err := cfg.UnmarshalKey(fmt.Sprintf("%s.%s", RootPluginName, PluginName), &p.cfg)
	if err != nil {
		return errors.E(op, err)
	}

	p.cfg.InitDefaults()

	// validate
	err = p.cfg.Valid()
	if err != nil {
		return errors.E(op, err)
	}
	// sub-plugin was initialized
	p.initialized = true
	return nil
}

func (p *Plugin) Provides() []interface{} {
	return []interface{}{
		p.ProvideServer,
	}
}

func (p *Plugin) ProvideServer() (*Server, error) {
	// if there is no ssl section
	// return nil config
	if !p.initialized {
		return nil, nil
	}
	p.server = p.initSSL()
	if p.cfg.RootCA != "" {
		err := p.appendRootCa()
		if err != nil {
			return nil, err
		}
	}
	l, err := utils.CreateListener(p.cfg.Address)
	if err != nil {
		return nil, err
	}
	s := &Server{
		p.server,
		l,
		p.redirect,
		p.cfg.Cert,
		p.cfg.Key,
	}

	return s, nil
}

//go:inline
func (p *Plugin) redirect(w http.ResponseWriter, r *http.Request) {
	target := &url.URL{
		Scheme: HTTPSScheme,
		// host or host:port
		Host:     p.tlsAddr(r.Host, false),
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	http.Redirect(w, r, target.String(), http.StatusPermanentRedirect)
}

// append RootCA to the https server TLS config
func (p *Plugin) appendRootCa() error {
	const op = errors.Op("http_plugin_append_root_ca")
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	CA, err := os.ReadFile(p.cfg.RootCA)
	if err != nil {
		return err
	}

	// should append our CA cert
	ok := rootCAs.AppendCertsFromPEM(CA)
	if !ok {
		return errors.E(op, errors.Str("could not append Certs from PEM"))
	}
	// disable "G402 (CWE-295): TLS MinVersion too low. (Confidence: HIGH, Severity: HIGH)"
	// #nosec G402
	cfg := &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            rootCAs,
	}
	p.server.TLSConfig = cfg

	return nil
}

// Init https server
func (p *Plugin) initSSL() *http.Server {
	var topCipherSuites []uint16
	var defaultCipherSuitesTLS13 []uint16

	hasGCMAsmAMD64 := cpu.X86.HasAES && cpu.X86.HasPCLMULQDQ
	hasGCMAsmARM64 := cpu.ARM64.HasAES && cpu.ARM64.HasPMULL
	// Keep in sync with crypto/aes/cipher_s390x.go.
	hasGCMAsmS390X := cpu.S390X.HasAES && cpu.S390X.HasAESCBC && cpu.S390X.HasAESCTR && (cpu.S390X.HasGHASH || cpu.S390X.HasAESGCM)

	hasGCMAsm := hasGCMAsmAMD64 || hasGCMAsmARM64 || hasGCMAsmS390X

	if hasGCMAsm {
		// If AES-GCM hardware is provided then priorities AES-GCM
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

	sslServer := &http.Server{
		Addr: p.tlsAddr(p.cfg.Address, true),
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

	return sslServer
}

// tlsAddr replaces listen or host port with port configured by SSLConfig config.
func (p *Plugin) tlsAddr(host string, forcePort bool) string {
	// remove current forcePort first
	host = strings.Split(host, ":")[0]

	if forcePort || p.cfg.port != 443 {
		host = fmt.Sprintf("%s:%v", host, p.cfg.port)
	}

	return host
}

func (p *Plugin) Name() string {
	return PluginName
}

// Available interface implementation
func (p *Plugin) Available() {}
