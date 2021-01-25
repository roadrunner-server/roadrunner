package gzip

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
)

const PluginName = "gzip"

type Gzip struct{}

// needed for the Endure
func (g *Gzip) Init() error {
	return nil
}

func (g *Gzip) Middleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gziphandler.GzipHandler(next).ServeHTTP(w, r)
	}
}

func (g *Gzip) Name() string {
	return PluginName
}
