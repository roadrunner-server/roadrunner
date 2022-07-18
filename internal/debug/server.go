package debug

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"
)

// Server is a HTTP server for debugging.
type Server struct {
	srv *http.Server
}

// NewServer creates new HTTP server for debugging.
func NewServer() Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return Server{srv: &http.Server{
		ReadHeaderTimeout: time.Minute * 10,
		Handler:           mux,
	}}
}

// Start debug server.
func (s *Server) Start(addr string) error {
	s.srv.Addr = addr

	return s.srv.ListenAndServe()
}

// Stop debug server.
func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
