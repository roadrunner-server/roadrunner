package headers

import (
	rrhttp "github.com/spiral/roadrunner/service/http"
	"net/http"
	"strconv"
)

// ID contains default service name.
const ID = "headers"

// Service serves headers files. Potentially convert into middleware?
type Service struct {
	// server configuration (location, forbidden files and etc)
	cfg *Config
}

// Init must return configure service and return true if service hasStatus enabled. Must return error in case of
// misconfiguration. Services must not be used without proper configuration pushed first.
func (s *Service) Init(cfg *Config, r *rrhttp.Service) (bool, error) {
	if r == nil {
		return false, nil
	}

	s.cfg = cfg
	r.AddMiddleware(s.middleware)

	return true, nil
}

// middleware must return true if request/response pair is handled within the middleware.
func (s *Service) middleware(f http.HandlerFunc) http.HandlerFunc {
	// Define the http.HandlerFunc
	return func(w http.ResponseWriter, r *http.Request) {

		if s.cfg.Request != nil {
			for k, v := range s.cfg.Request {
				r.Header.Add(k, v)
			}
		}

		if s.cfg.Response != nil {
			for k, v := range s.cfg.Response {
				w.Header().Set(k, v)
			}
		}

		if s.cfg.CORS != nil {
			if r.Method == http.MethodOptions {
				s.preflightRequest(w, r)
				return
			}

			s.corsHeaders(w, r)
		}

		f(w, r)
	}
}

// configure OPTIONS response
func (s *Service) preflightRequest(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()

	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")

	if s.cfg.CORS.AllowedOrigin != "" {
		headers.Set("Access-Control-Allow-Origin", s.cfg.CORS.AllowedOrigin)
	}

	if s.cfg.CORS.AllowedHeaders != "" {
		headers.Set("Access-Control-Allow-Headers", s.cfg.CORS.AllowedHeaders)
	}

	if s.cfg.CORS.AllowedMethods != "" {
		headers.Set("Access-Control-Allow-Methods", s.cfg.CORS.AllowedMethods)
	}

	if s.cfg.CORS.AllowCredentials != nil {
		headers.Set("Access-Control-Allow-Credentials", strconv.FormatBool(*s.cfg.CORS.AllowCredentials))
	}

	if s.cfg.CORS.MaxAge > 0 {
		headers.Set("Access-Control-Max-Age", strconv.Itoa(s.cfg.CORS.MaxAge))
	}

	w.WriteHeader(http.StatusOK)
}

// configure CORS headers
func (s *Service) corsHeaders(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()

	headers.Add("Vary", "Origin")

	if s.cfg.CORS.AllowedOrigin != "" {
		headers.Set("Access-Control-Allow-Origin", s.cfg.CORS.AllowedOrigin)
	}

	if s.cfg.CORS.AllowedHeaders != "" {
		headers.Set("Access-Control-Allow-Headers", s.cfg.CORS.AllowedHeaders)
	}

	if s.cfg.CORS.ExposedHeaders != "" {
		headers.Set("Access-Control-Expose-Headers", s.cfg.CORS.ExposedHeaders)
	}

	if s.cfg.CORS.AllowCredentials != nil {
		headers.Set("Access-Control-Allow-Credentials", strconv.FormatBool(*s.cfg.CORS.AllowCredentials))
	}
}
