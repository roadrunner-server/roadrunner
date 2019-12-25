package gzip

import (
	rrhttp "github.com/spiral/roadrunner/service/http"
	"github.com/NYTimes/gziphandler"
	"net/http"
)

// ID contains default service name.
const ID = "gzip"

type Service struct {
	cfg *Config
}

func (s *Service) Init(cfg *Config, r *rrhttp.Service) (bool, error) {
	s.cfg = cfg

	if s.cfg.Enable == false {
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
