package http

import (
	"net/http"
	"strconv"
	"github.com/spiral/roadrunner"
	"github.com/pkg/errors"
)

// Server serves http connections to underlying PHP application using PSR-7 protocol. Context will include request headers,
// parsed files and query, payload will include parsed form dataTree (if any).
type Server struct {
	cfg      *Config
	listener func(event int, ctx interface{})
	rr       *roadrunner.Server
}

// Listen attaches pool event watcher.
func (s *Server) Listen(l func(event int, ctx interface{})) {
	s.listener = l
}

// Handle serve using PSR-7 requests passed to underlying application. Attempts to serve static files first if enabled.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// validating request size
	if s.cfg.MaxRequest != 0 {
		if length := r.Header.Get("content-length"); length != "" {
			if size, err := strconv.ParseInt(length, 10, 64); err != nil {
				s.handleError(w, r, err)
				return
			} else if size > s.cfg.MaxRequest {
				s.handleError(w, r, errors.New("request body max size is exceeded"))
				return
			}
		}
	}

	req, err := NewRequest(r, s.cfg.Uploads)
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	if err = req.Open(); err != nil {
		s.handleError(w, r, err)
		return
	}
	defer req.Close()

	p, err := req.Payload()
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	rsp, err := s.rr.Exec(p)
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	resp, err := NewResponse(rsp)
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	resp.Write(w)
}

// handleError sends error
func (s *Server) handleError(w http.ResponseWriter, r *http.Request, err error) {
	s.throw(2332323, err)

	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}

// throw invokes event srv if any.
func (s *Server) throw(event int, ctx interface{}) {
	if s.listener != nil {
		s.listener(event, ctx)
	}
}
