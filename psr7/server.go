package psr7

import (
	"github.com/spiral/roadrunner"
	"net/http"
	"strings"
	"path"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"encoding/json"
)

var (
	excludeFiles = []string{".php", ".htaccess"}
)

// Configures http rr
type Config struct {
	// ServeStatic enables static file serving from desired root directory.
	ServeStatic bool

	// Root directory, required when ServeStatic set to true.
	Root string
}

// Server serves http connections to underlying PHP application using PSR-7 protocol. Context will include request headers,
// parsed files and query, payload will include parsed form data (if any) - todo: do we need to do that?.
type Server struct {
	cfg  Config
	root http.Dir
	rr   *roadrunner.Server
}

// NewServer returns new instance of Server PSR7 server.
func NewServer(cfg Config, server *roadrunner.Server) *Server {
	h := &Server{cfg: cfg, rr: server}
	if cfg.ServeStatic {
		h.root = http.Dir(h.cfg.Root)
	}

	return h
}

// ServeHTTP serve using PSR-7 requests passed to underlying application.
func (svr *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) () {
	if svr.cfg.ServeStatic && svr.serveStatic(w, r) {
		// server always attempt to serve static files first
		return
	}

	req, err := ParseRequest(r)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error())) //todo: better errors
		return
	}
	defer req.Close()

	rsp, err := svr.rr.Exec(req.Payload())
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error())) //todo: better errors
		return
	}

	resp := &Response{}
	if err = json.Unmarshal(rsp.Context, resp); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error())) //todo: better errors
		return
	}

	resp.Write(w)
	w.Write(rsp.Body)
}

// serveStatic attempts to serve static file and returns true in case of success, will return false in case if file not
// found, not allowed or on read error.
func (svr *Server) serveStatic(w http.ResponseWriter, r *http.Request) bool {
	fpath := r.URL.Path
	if !strings.HasPrefix(fpath, "/") {
		fpath = "/" + fpath
	}
	fpath = path.Clean(fpath)

	if svr.excluded(fpath) {
		logrus.Warningf("attempt to access forbidden file %s", fpath)
		return false
	}

	f, err := svr.root.Open(fpath)
	if err != nil {
		if !os.IsNotExist(err) {
			// rr or access error
			logrus.Error(err)
		}

		return false
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		// rr error
		logrus.Error(err)
		return false
	}

	if d.IsDir() {
		// we are not serving directories
		return false
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	return true
}

// excluded returns true if file has forbidden extension.
func (svr *Server) excluded(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, exl := range excludeFiles {
		if ext == exl {
			return true
		}
	}

	return false
}
