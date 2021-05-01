package http

import (
	"net/http"
	"net/http/fcgi"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/utils"
	"golang.org/x/net/http2"
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
	//if s.https == nil {
	//	return
	//}
	//const op = errors.Op("http_plugin_serve_https")
	//if len(s.mdwr) > 0 {
	//	applyMiddlewares(s.https, s.mdwr, s.cfg.Middleware, s.log)
	//}
	//l, err := utils.CreateListener(s.cfg.SSLConfig.Address)
	//if err != nil {
	//	errCh <- errors.E(op, err)
	//	return
	//}
	//
	//err = s.https.ServeTLS(
	//	l,
	//	s.cfg.SSLConfig.Cert,
	//	s.cfg.SSLConfig.Key,
	//)
	//
	//if err != nil && err != http.ErrServerClosed {
	//	errCh <- errors.E(op, err)
	//	return
	//}
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

// https://golang.org/pkg/net/http/#Hijacker
//go:inline
func headerContainsUpgrade(r *http.Request) bool {
	if _, ok := r.Header["Upgrade"]; ok {
		return true
	}
	return false
}

// init http/2 server
func (s *Plugin) initHTTP2() error {
	return http2.ConfigureServer(s.https, &http2.Server{
		MaxConcurrentStreams: s.cfg.HTTP2Config.MaxConcurrentStreams,
	})
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
