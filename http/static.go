package http

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path"
	"strings"
)

// staticServer serves static files
type staticServer struct {
	cfg  *FsConfig
	root http.Dir
}

// serve attempts to serve static file and returns true in case of success, will return false in case if file not
// found, not allowed or on read error.
func (svr *staticServer) serve(w http.ResponseWriter, r *http.Request) bool {
	fPath := r.URL.Path
	if !strings.HasPrefix(fPath, "/") {
		fPath = "/" + fPath
	}
	fPath = path.Clean(fPath)

	if svr.cfg.Forbids(fPath) {
		logrus.Warningf("attempt to access forbidden file %s", fPath) // todo: better logs
		return false
	}

	f, err := svr.root.Open(fPath)
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

		// todo: do i need it, bypass log?

		return false
	}

	if d.IsDir() {
		// do not serve directories
		return false
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	return true
}
