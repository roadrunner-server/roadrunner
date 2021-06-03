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

func (p *Plugin) serveHTTP(errCh chan error) {
	if p.http == nil {
		return
	}
	const op = errors.Op("serveHTTP")

	if len(p.mdwr) > 0 {
		applyMiddlewares(p.http, p.mdwr, p.cfg.Middleware, p.log)
	}
	l, err := utils.CreateListener(p.cfg.Address)
	if err != nil {
		errCh <- errors.E(op, err)
		return
	}

	err = p.http.Serve(l)
	if err != nil && err != http.ErrServerClosed {
		errCh <- errors.E(op, err)
		return
	}
}

func (p *Plugin) serveHTTPS(errCh chan error) {
	if p.https == nil {
		return
	}
	const op = errors.Op("serveHTTPS")
	if len(p.mdwr) > 0 {
		applyMiddlewares(p.https, p.mdwr, p.cfg.Middleware, p.log)
	}
	l, err := utils.CreateListener(p.cfg.SSLConfig.Address)
	if err != nil {
		errCh <- errors.E(op, err)
		return
	}

	err = p.https.ServeTLS(
		l,
		p.cfg.SSLConfig.Cert,
		p.cfg.SSLConfig.Key,
	)

	if err != nil && err != http.ErrServerClosed {
		errCh <- errors.E(op, err)
		return
	}
}

// serveFCGI starts FastCGI server.
func (p *Plugin) serveFCGI(errCh chan error) {
	if p.fcgi == nil {
		return
	}
	const op = errors.Op("serveFCGI")

	if len(p.mdwr) > 0 {
		applyMiddlewares(p.fcgi, p.mdwr, p.cfg.Middleware, p.log)
	}

	l, err := utils.CreateListener(p.cfg.FCGIConfig.Address)
	if err != nil {
		errCh <- errors.E(op, err)
		return
	}

	err = fcgi.Serve(l, p.fcgi.Handler)
	if err != nil && err != http.ErrServerClosed {
		errCh <- errors.E(op, err)
		return
	}
}

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

// https://golang.org/pkg/net/http/#Hijacker
//go:inline
func headerContainsUpgrade(r *http.Request) bool {
	if _, ok := r.Header["Upgrade"]; ok {
		return true
	}
	return false
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

	CA, err := os.ReadFile(p.cfg.SSLConfig.RootCA)
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
	p.http.TLSConfig = cfg

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
		Addr:     p.tlsAddr(p.cfg.Address, true),
		Handler:  p,
		ErrorLog: p.stdLog,
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
func (p *Plugin) initHTTP2() error {
	return http2.ConfigureServer(p.https, &http2.Server{
		MaxConcurrentStreams: p.cfg.HTTP2Config.MaxConcurrentStreams,
	})
}

// tlsAddr replaces listen or host port with port configured by SSLConfig config.
func (p *Plugin) tlsAddr(host string, forcePort bool) string {
	// remove current forcePort first
	host = strings.Split(host, ":")[0]

	if forcePort || p.cfg.SSLConfig.Port != 443 {
		host = fmt.Sprintf("%s:%v", host, p.cfg.SSLConfig.Port)
	}

	return host
}

// static plugin name
const static string = "static"

func applyMiddlewares(server *http.Server, middlewares map[string]Middleware, order []string, log logger.Logger) {
	for i := len(order) - 1; i >= 0; i-- {
		// set static last in the row
		if order[i] == static {
			continue
		}
		if mdwr, ok := middlewares[order[i]]; ok {
			server.Handler = mdwr.Middleware(server.Handler)
		} else {
			log.Warn("requested middleware does not exist", "requested", order[i])
		}
	}

	// set static if exists
	if mdwr, ok := middlewares[static]; ok {
		server.Handler = mdwr.Middleware(server.Handler)
	}
}
