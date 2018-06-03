package http

import (
	"github.com/spiral/roadrunner"
	"net/http"
	"strconv"
	"errors"
	"github.com/sirupsen/logrus"
)

// Service serves http connections to underlying PHP application using PSR-7 protocol. Context will include request headers,
// parsed files and query, payload will include parsed form dataTree (if any).
type Server struct {
	cfg    *Config
	static *staticServer
	rr     *roadrunner.Server
}

// NewServer returns new instance of HTTP PSR7 server.
func NewServer(cfg *Config, server *roadrunner.Server) *Server {
	h := &Server{cfg: cfg, rr: server}

	if cfg.ServeStatic {
		h.static = &staticServer{root: http.Dir(h.cfg.Root)}
	}

	return h
}

// ServeHTTP serve using PSR-7 requests passed to underlying application. Attempts to serve static files first if enabled.
func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) () {
	if srv.cfg.ServeStatic && srv.static.serve(w, r) {
		return
	}

	// validating request size
	if srv.cfg.MaxRequest != 0 {
		if length := r.Header.Get("content-length"); length != "" {
			if size, err := strconv.ParseInt(length, 10, 64); err != nil {
				srv.sendError(w, r, err)
				return
			} else if size > srv.cfg.MaxRequest {
				srv.sendError(w, r, errors.New("request body max size is exceeded"))
				return
			}
		}
	}

	req, err := NewRequest(r)
	if err != nil {
		srv.sendError(w, r, err)
		return
	}

	if err = req.Open(srv.cfg); err != nil {
		srv.sendError(w, r, err)
		return
	}
	defer req.Close()

	p, err := req.Payload()
	if err != nil {
		srv.sendError(w, r, err)
		return
	}

	rsp, err := srv.rr.Exec(p)
	if err != nil {
		srv.sendError(w, r, err)
		return
	}

	resp, err := NewResponse(rsp)
	if err != nil {
		srv.sendError(w, r, err)
		return
	}

	resp.Write(w)
}

// sendError sends error
func (srv *Server) sendError(w http.ResponseWriter, r *http.Request, err error) {
	if _, job := err.(roadrunner.JobError); !job {
		logrus.Error(err)
	}

	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}
