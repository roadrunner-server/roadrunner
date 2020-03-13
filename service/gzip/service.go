package gzip

import (
	"errors"
	"github.com/NYTimes/gziphandler"
	rrhttp "github.com/spiral/roadrunner/service/http"
	"net/http"
)

// ID contains default service name.
const ID = "gzip"
var httpNotInitialized = errors.New("http service should be defined properly in config")

type Service struct {
	cfg *Config
}

func (s *Service) Init(cfg *Config, r *rrhttp.Service) (bool, error) {
	if r == nil {
		return false, httpNotInitialized
	}
	s.cfg = cfg

	if !s.cfg.Enable {
		return false, nil
	}

	r.AddMiddleware(s.middleware)

	return true, nil
}

func (s *Service) middleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gziphandler.GzipHandler(f).ServeHTTP(w, r)
	}
}
