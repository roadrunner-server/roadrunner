package http

import (
	"net/http"

	"github.com/spiral/roadrunner-plugins/config"
)

type PluginMiddleware struct {
	config config.Configurer
}

func (p *PluginMiddleware) Init(cfg config.Configurer) error {
	p.config = cfg
	return nil
}

func (p *PluginMiddleware) Middleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/halt" {
			w.WriteHeader(500)
			_, err := w.Write([]byte("halted"))
			if err != nil {
				panic("error writing the data to the http reply")
			}
		} else {
			next.ServeHTTP(w, r)
		}
	}
}

func (p *PluginMiddleware) Name() string {
	return "pluginMiddleware"
}

type PluginMiddleware2 struct {
	config config.Configurer
}

func (p *PluginMiddleware2) Init(cfg config.Configurer) error {
	p.config = cfg
	return nil
}

func (p *PluginMiddleware2) Middleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/boom" {
			w.WriteHeader(555)
			_, err := w.Write([]byte("boom"))
			if err != nil {
				panic("error writing the data to the http reply")
			}
		} else {
			next.ServeHTTP(w, r)
		}
	}
}

func (p *PluginMiddleware2) Name() string {
	return "pluginMiddleware2"
}
