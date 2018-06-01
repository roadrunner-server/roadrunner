package http

import (
	"github.com/spiral/roadrunner"
	"net/http"
)

// Configures RoadRunner HTTP server.
type Config struct {
	// serve enables static file serving from desired root directory.
	ServeStatic bool

	// Root directory, required when serve set to true.
	Root string

	// UploadsDir contains name of temporary directory to store uploaded files passed to underlying PHP process.
	UploadsDir string
}

// Server serves http connections to underlying PHP application using PSR-7 protocol. Context will include request headers,
// parsed files and query, payload will include parsed form dataTree (if any).
type Server struct {
	cfg    Config
	static *staticServer
	rr     *roadrunner.Server
}

// NewServer returns new instance of HTTP PSR7 server.
func NewServer(cfg Config, server *roadrunner.Server) *Server {
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

	req, err := NewRequest(r)
	if err != nil {
		srv.sendError(w, r, err)
		return
	}

	if err = req.OpenUploads(srv.cfg.UploadsDir); err != nil {
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
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}
