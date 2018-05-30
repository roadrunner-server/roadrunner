package server

import (
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/server/psr7"
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
type HTTPConfig struct {
	// ServeStatic enables static file serving from desired root directory.
	ServeStatic bool

	// Root directory, required when ServeStatic set to true.
	Root string
}

// HTTP serves http connections to underlying PHP application using PSR-7 protocol. Context will include request headers,
// parsed files and query, payload will include parsed form data (if any) - todo: do we need to do that?.
type HTTP struct {
	cfg  HTTPConfig
	root http.Dir
	rr   *roadrunner.Server
}

// NewHTTP returns new instance of HTTP PSR7 server.
func NewHTTP(cfg HTTPConfig, server *roadrunner.Server) *HTTP {
	h := &HTTP{cfg: cfg, rr: server}
	if cfg.ServeStatic {
		h.root = http.Dir(h.cfg.Root)
	}

	return h
}

// ServeHTTP serve using PSR-7 requests passed to underlying application.
func (h *HTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) () {
	if h.cfg.ServeStatic && h.serveStatic(w, r) {
		// server always attempt to serve static files first
		return
	}

	req, err := psr7.ParseRequest(r)
	if err != nil {
		w.Write([]byte(err.Error())) //todo: better errors
		w.WriteHeader(500)
		return
	}
	defer req.Close()

	rsp, err := h.rr.Exec(req.Payload())
	if err != nil {
		w.Write([]byte(err.Error())) //todo: better errors
		w.WriteHeader(500)
		return
	}

	resp := &psr7.Response{}
	if err = json.Unmarshal(rsp.Context, resp); err != nil {
		w.Write([]byte(err.Error())) //todo: better errors
		w.WriteHeader(500)
		return
	}

	resp.Write(w)
	w.Write(rsp.Body)
}

// serveStatic attempts to serve static file and returns true in case of success, will return false in case if file not
// found, not allowed or on read error.
func (h *HTTP) serveStatic(w http.ResponseWriter, r *http.Request) bool {
	fpath := r.URL.Path
	if !strings.HasPrefix(fpath, "/") {
		fpath = "/" + fpath
	}
	fpath = path.Clean(fpath)

	if h.excluded(fpath) {
		logrus.Warningf("attempt to access forbidden file %s", fpath)
		return false
	}

	f, err := h.root.Open(fpath)
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
func (h *HTTP) excluded(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, exl := range excludeFiles {
		if ext == exl {
			return true
		}
	}

	return false
}
