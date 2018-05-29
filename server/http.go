package server

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

type Request struct {
	Protocol string            `json:"protocol"`
	Uri      string            `json:"uri"`
	Method   string            `json:"method"`
	Headers  http.Header       `json:"headers"`
	Cookies  map[string]string `json:"cookies"`
	RawQuery string            `json:"rawQuery"`
}

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
		// serving static files
		return
	}

	// WHAT TO PUT TO BODY?
	p, err := h.buildPayload(r)
	rsp, err := h.rr.Exec(p)

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	// wrapping the response

	w.Header().Add("content-type", "text/html;charset=UTF-8")
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

	if isForbidden(fpath) {
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

// todo: add files support
func (h *HTTP) buildPayload(r *http.Request) (*roadrunner.Payload, error) {
	request := Request{
		Protocol: r.Proto,
		Uri:      r.URL.String(),
		Method:   r.Method,
		Headers:  r.Header,
		Cookies:  make(map[string]string),
		RawQuery: r.URL.RawQuery,
	}

	r.ParseMultipartForm(1000000000)

	logrus.Print(r.MultipartForm.Value)
	logrus.Print(r.MultipartForm.File["kkk"][0].Header)
	logrus.Print(r.MultipartForm.File["kkk"][0].Filename)
	logrus.Print(r.Form)

	// cookies
	for _, c := range r.Cookies() {
		request.Cookies[c.Name] = c.Value
	}

	data, _ := json.Marshal(request)

	logrus.Info(string(data))

	return &roadrunner.Payload{
		Context: data,
		Body:    []byte("lol"),
	}, nil
}

// isForbidden returns true if file has forbidden extension.
func isForbidden(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, exl := range excludeFiles {
		if ext == exl {
			return true
		}
	}

	return false
}
