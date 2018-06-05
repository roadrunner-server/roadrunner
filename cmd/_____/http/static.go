package http

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	forbiddenFiles = []string{".php", ".htaccess"}
)

// staticServer serves static files
type staticServer struct {
	root http.Dir
}

// serve attempts to serve static file and returns true in case of success, will return false in case if file not
// found, not allowed or on read error.
func (svr *staticServer) serve(w http.ResponseWriter, r *http.Request) bool {
	fpath := r.URL.Path
	if !strings.HasPrefix(fpath, "/") {
		fpath = "/" + fpath
	}
	fpath = path.Clean(fpath)

	if svr.forbidden(fpath) {
		logrus.Warningf("attempt to access forbidden file %s", fpath) // todo: better logs
		return false
	}

	f, err := svr.root.Open(fpath)
	if err != nil {
		if !os.IsNotExist(err) {
			logrus.Error(err) //todo: rr or access error
		}

		return false
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		logrus.Error(err) //todo: rr or access error
		return false
	}

	if d.IsDir() {
		// do not serve directories
		return false
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	return true
}

// forbidden returns true if file has forbidden extension.
func (svr *staticServer) forbidden(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, exl := range forbiddenFiles {
		if ext == exl {
			return true
		}
	}

	return false
}
