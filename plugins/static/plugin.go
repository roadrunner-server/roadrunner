package static

import (
	"net/http"
	"path"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

// ID contains default service name.
const PluginName = "static"

const RootPluginName = "http"

// Plugin serves static files. Potentially convert into middleware?
type Plugin struct {
	// server configuration (location, forbidden files and etc)
	cfg *Config

	log logger.Logger

	// root is initiated http directory
	root http.Dir
}

// Init must return configure service and return true if service hasStatus enabled. Must return error in case of
// misconfiguration. Services must not be used without proper configuration pushed first.
func (s *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("static_plugin_init")
	if !cfg.Has(RootPluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(RootPluginName, &s.cfg)
	if err != nil {
		return errors.E(op, errors.Disabled, err)
	}

	if s.cfg.Static == nil {
		return errors.E(op, errors.Disabled)
	}

	s.log = log
	s.root = http.Dir(s.cfg.Static.Dir)

	err = s.cfg.Valid()
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (s *Plugin) Name() string {
	return PluginName
}

// middleware must return true if request/response pair is handled within the middleware.
func (s *Plugin) Middleware(next http.Handler) http.HandlerFunc {
	// Define the http.HandlerFunc
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.Static.Request != nil {
			for k, v := range s.cfg.Static.Request {
				r.Header.Add(k, v)
			}
		}

		if s.cfg.Static.Response != nil {
			for k, v := range s.cfg.Static.Response {
				w.Header().Set(k, v)
			}
		}

		if !s.handleStatic(w, r) {
			next.ServeHTTP(w, r)
		}
	}
}

func (s *Plugin) handleStatic(w http.ResponseWriter, r *http.Request) bool {
	fPath := path.Clean(r.URL.Path)

	if s.cfg.AlwaysForbid(fPath) {
		return false
	}

	f, err := s.root.Open(fPath)
	if err != nil {
		if s.cfg.AlwaysServe(fPath) {
			w.WriteHeader(404)
			return true
		}

		return false
	}
	defer func() {
		err = f.Close()
		if err != nil {
			s.log.Error("file closing error", "error", err)
		}
	}()

	d, err := f.Stat()
	if err != nil {
		return false
	}

	// do not serve directories
	if d.IsDir() {
		return false
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	return true
}
