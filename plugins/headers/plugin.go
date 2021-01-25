package headers

import (
	"net/http"
	"strconv"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
)

// ID contains default service name.
const PluginName = "headers"
const RootPluginName = "http"

// Service serves headers files. Potentially convert into middleware?
type Plugin struct {
	// server configuration (location, forbidden files and etc)
	cfg *Config
}

// Init must return configure service and return true if service hasStatus enabled. Must return error in case of
// misconfiguration. Services must not be used without proper configuration pushed first.
func (s *Plugin) Init(cfg config.Configurer) error {
	const op = errors.Op("headers_plugin_init")
	if !cfg.Has(RootPluginName) {
		return errors.E(op, errors.Disabled)
	}
	err := cfg.UnmarshalKey(RootPluginName, &s.cfg)
	if err != nil {
		return errors.E(op, errors.Disabled, err)
	}

	if s.cfg.Headers == nil {
		return errors.E(op, errors.Disabled)
	}

	return nil
}

// middleware must return true if request/response pair is handled within the middleware.
func (s *Plugin) Middleware(next http.Handler) http.HandlerFunc {
	// Define the http.HandlerFunc
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.Headers.Request != nil {
			for k, v := range s.cfg.Headers.Request {
				r.Header.Add(k, v)
			}
		}

		if s.cfg.Headers.Response != nil {
			for k, v := range s.cfg.Headers.Response {
				w.Header().Set(k, v)
			}
		}

		if s.cfg.Headers.CORS != nil {
			if r.Method == http.MethodOptions {
				s.preflightRequest(w)
				return
			}
			s.corsHeaders(w)
		}

		next.ServeHTTP(w, r)
	}
}

func (s *Plugin) Name() string {
	return PluginName
}

// configure OPTIONS response
func (s *Plugin) preflightRequest(w http.ResponseWriter) {
	headers := w.Header()

	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")

	if s.cfg.Headers.CORS.AllowedOrigin != "" {
		headers.Set("Access-Control-Allow-Origin", s.cfg.Headers.CORS.AllowedOrigin)
	}

	if s.cfg.Headers.CORS.AllowedHeaders != "" {
		headers.Set("Access-Control-Allow-Headers", s.cfg.Headers.CORS.AllowedHeaders)
	}

	if s.cfg.Headers.CORS.AllowedMethods != "" {
		headers.Set("Access-Control-Allow-Methods", s.cfg.Headers.CORS.AllowedMethods)
	}

	if s.cfg.Headers.CORS.AllowCredentials != nil {
		headers.Set("Access-Control-Allow-Credentials", strconv.FormatBool(*s.cfg.Headers.CORS.AllowCredentials))
	}

	if s.cfg.Headers.CORS.MaxAge > 0 {
		headers.Set("Access-Control-Max-Age", strconv.Itoa(s.cfg.Headers.CORS.MaxAge))
	}

	w.WriteHeader(http.StatusOK)
}

// configure CORS headers
func (s *Plugin) corsHeaders(w http.ResponseWriter) {
	headers := w.Header()

	headers.Add("Vary", "Origin")

	if s.cfg.Headers.CORS.AllowedOrigin != "" {
		headers.Set("Access-Control-Allow-Origin", s.cfg.Headers.CORS.AllowedOrigin)
	}

	if s.cfg.Headers.CORS.AllowedHeaders != "" {
		headers.Set("Access-Control-Allow-Headers", s.cfg.Headers.CORS.AllowedHeaders)
	}

	if s.cfg.Headers.CORS.ExposedHeaders != "" {
		headers.Set("Access-Control-Expose-Headers", s.cfg.Headers.CORS.ExposedHeaders)
	}

	if s.cfg.Headers.CORS.AllowCredentials != nil {
		headers.Set("Access-Control-Allow-Credentials", strconv.FormatBool(*s.cfg.Headers.CORS.AllowCredentials))
	}
}
