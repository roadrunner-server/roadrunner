package static

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

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

	// file extensions which are allowed to be served
	allowedExtensions map[string]struct{}

	// file extensions which are forbidden to be served
	forbiddenExtensions map[string]struct{}

	alwaysServe map[string]struct{}
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

	// create 2 hashmaps with the allowed and forbidden file extensions
	s.allowedExtensions = make(map[string]struct{}, len(s.cfg.Static.Allow))
	s.forbiddenExtensions = make(map[string]struct{}, len(s.cfg.Static.Forbid))
	s.alwaysServe = make(map[string]struct{}, len(s.cfg.Static.Always))

	for i := 0; i < len(s.cfg.Static.Forbid); i++ {
		s.forbiddenExtensions[s.cfg.Static.Forbid[i]] = struct{}{}
	}

	for i := 0; i < len(s.cfg.Static.Allow); i++ {
		s.forbiddenExtensions[s.cfg.Static.Allow[i]] = struct{}{}
	}

	// check if any forbidden items presented in the allowed
	// if presented, delete such items from allowed
	for k := range s.forbiddenExtensions {
		if _, ok := s.allowedExtensions[k]; ok {
			delete(s.allowedExtensions, k)
		}
	}

	for i := 0; i < len(s.cfg.Static.Always); i++ {
		s.alwaysServe[s.cfg.Static.Always[i]] = struct{}{}
	}

	// at this point we have distinct allowed and forbidden hashmaps, also with alwaysServed

	return nil
}

func (s *Plugin) Name() string {
	return PluginName
}

// Middleware must return true if request/response pair is handled within the middleware.
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

		fPath := path.Clean(r.URL.Path)
		ext := strings.ToLower(path.Ext(fPath))

		// check that file is in forbidden list
		if _, ok := s.forbiddenExtensions[ext]; ok {
			http.Error(w, "file is forbidden", 404)
			return
		}

		f, err := s.root.Open(fPath)
		if err != nil {
			// if we should always serve files with some extensions
			// show error to the user and invoke next middleware
			if _, ok := s.alwaysServe[ext]; ok {
				//http.Error(w, err.Error(), 404)
				w.WriteHeader(404)
				next.ServeHTTP(w, r)
				return
			}
			// else, return with error
			http.Error(w, err.Error(), 404)
			return
		}

		defer func() {
			err = f.Close()
			if err != nil {
				s.log.Error("file close error", "error", err)
			}
		}()

		// here we know, that file extension is not in the AlwaysServe and file exists
		// (or by some reason, there is no error from the http.Open method)

		// if we have some allowed extensions, we should check them
		if len(s.allowedExtensions) > 0 {
			if _, ok := s.allowedExtensions[ext]; ok {
				d, err := s.check(f)
				if err != nil {
					http.Error(w, err.Error(), 404)
					return
				}

				http.ServeContent(w, r, d.Name(), d.ModTime(), f)
			}
			// otherwise we guess, that all file extensions are allowed
		} else {
			d, err := s.check(f)
			if err != nil {
				http.Error(w, err.Error(), 404)
				return
			}

			http.ServeContent(w, r, d.Name(), d.ModTime(), f)
		}

		next.ServeHTTP(w, r)
	}
}

func (s *Plugin) check(f http.File) (fs.FileInfo, error) {
	const op = errors.Op("http_file_check")
	d, err := f.Stat()
	if err != nil {
		return nil, err
	}

	// do not serve directories
	if d.IsDir() {
		return nil, errors.E(op, errors.Str("directory path provided, should be path to the file"))
	}

	return d, nil
}
