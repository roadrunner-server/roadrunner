package http

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/fcgi"
	"net/url"
	"os"
	"strings"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/utils"
	"golang.org/x/net/http2"
	"golang.org/x/sys/cpu"
)

func (s *Plugin) serveHTTP(errCh chan error) {
	if s.http == nil {
		return
	}
	const op = errors.Op("http_plugin_serve_http")

	if len(s.mdwr) > 0 {
		applyMiddlewares(s.http, s.mdwr, s.cfg.Middleware, s.log)
	}
	l, err := utils.CreateListener(s.cfg.Address)
	if err != nil {
		errCh <- errors.E(op, err)
		return
	}

	err = s.http.Serve(l)
	if err != nil && err != http.ErrServerClosed {
		errCh <- errors.E(op, err)
		return
	}
}

func (s *Plugin) serveHTTPS(errCh chan error) {
	if s.https == nil {
		return
	}
	const op = errors.Op("http_plugin_serve_https")
	if len(s.mdwr) > 0 {
		applyMiddlewares(s.https, s.mdwr, s.cfg.Middleware, s.log)
	}
	l, err := utils.CreateListener(s.cfg.SSLConfig.Address)
	if err != nil {
		errCh <- errors.E(op, err)
		return
	}

	err = s.https.ServeTLS(
		l,
		s.cfg.SSLConfig.Cert,
		s.cfg.SSLConfig.Key,
	)

	if err != nil && err != http.ErrServerClosed {
		errCh <- errors.E(op, err)
		return
	}
}

// serveFCGI starts FastCGI server.
func (s *Plugin) serveFCGI(errCh chan error) {
	if s.fcgi == nil {
		return
	}
	const op = errors.Op("http_plugin_serve_fcgi")

	if len(s.mdwr) > 0 {
		applyMiddlewares(s.https, s.mdwr, s.cfg.Middleware, s.log)
	}

	l, err := utils.CreateListener(s.cfg.FCGIConfig.Address)
	if err != nil {
		errCh <- errors.E(op, err)
		return
	}

	err = fcgi.Serve(l, s.fcgi.Handler)
	if err != nil && err != http.ErrServerClosed {
		errCh <- errors.E(op, err)
		return
	}
}

func (s *Plugin) redirect(w http.ResponseWriter, r *http.Request) {
	target := &url.URL{
		Scheme: HTTPSScheme,
		// host or host:port
		Host:     s.tlsAddr(r.Host, false),
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	http.Redirect(w, r, target.String(), http.StatusPermanentRedirect)
}

// https://golang.org/pkg/net/http/#Hijacker
//go:inline
func headerContainsUpgrade(r *http.Request) bool {
	if _, ok := r.Header["Upgrade"]; ok {
		return true
	}
	return false
}

// append RootCA to the https server TLS config
func (s *Plugin) appendRootCa() error {
	const op = errors.Op("http_plugin_append_root_ca")
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	CA, err := os.ReadFile(s.cfg.SSLConfig.RootCA)
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
	s.http.TLSConfig = cfg

	return nil
}

// Init https server
func (s *Plugin) initSSL() *http.Server {
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
		Addr:     s.tlsAddr(s.cfg.Address, true),
		Handler:  s,
		ErrorLog: s.stdLog,
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

// init http/2 server
func (s *Plugin) initHTTP2() error {
	return http2.ConfigureServer(s.https, &http2.Server{
		MaxConcurrentStreams: s.cfg.HTTP2Config.MaxConcurrentStreams,
	})
}

// tlsAddr replaces listen or host port with port configured by SSLConfig config.
func (s *Plugin) tlsAddr(host string, forcePort bool) string {
	// remove current forcePort first
	host = strings.Split(host, ":")[0]

	if forcePort || s.cfg.SSLConfig.Port != 443 {
		host = fmt.Sprintf("%s:%v", host, s.cfg.SSLConfig.Port)
	}

	return host
}

func applyMiddlewares(server *http.Server, middlewares map[string]Middleware, order []string, log logger.Logger) {
	for i := len(order) - 1; i >= 0; i-- {
		if mdwr, ok := middlewares[order[i]]; ok {
			server.Handler = mdwr.Middleware(server.Handler)
		} else {
			log.Warn("requested middleware does not exist", "requested", order[i])
		}
	}
}
