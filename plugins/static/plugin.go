package static

import (
	"net/http"
	"path"
	"strings"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

// PluginName contains default service name.
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

	// init forbidden
	for i := 0; i < len(s.cfg.Static.Forbid); i++ {
		// skip empty lines
		if s.cfg.Static.Forbid[i] == "" {
			continue
		}
		s.forbiddenExtensions[s.cfg.Static.Forbid[i]] = struct{}{}
	}

	// init allowed
	for i := 0; i < len(s.cfg.Static.Allow); i++ {
		// skip empty lines
		if s.cfg.Static.Allow[i] == "" {
			continue
		}
		s.allowedExtensions[s.cfg.Static.Allow[i]] = struct{}{}
	}

	// check if any forbidden items presented in the allowed
	// if presented, delete such items from allowed
	for k := range s.forbiddenExtensions {
		delete(s.allowedExtensions, k)
	}

	// at this point we have distinct allowed and forbidden hashmaps, also with alwaysServed
	return nil
}

func (s *Plugin) Name() string {
	return PluginName
}

// Middleware must return true if request/response pair is handled within the middleware.
func (s *Plugin) Middleware(next http.Handler) http.Handler {
	// Define the http.HandlerFunc
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// do not allow paths like ../../resource
		// only specified folder and resources in it
		// https://lgtm.com/rules/1510366186013/
		if strings.Contains(r.URL.Path, "..") {
			w.WriteHeader(http.StatusForbidden)
			return
		}

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

		// first - create a proper file path
		fPath := path.Clean(r.URL.Path)
		ext := strings.ToLower(path.Ext(fPath))

		// check that file extension in the forbidden list
		if _, ok := s.forbiddenExtensions[ext]; ok {
			s.log.Debug("file extension is forbidden", "ext", ext)
			next.ServeHTTP(w, r)
			return
		}

		// if we have some allowed extensions, we should check them
		// if not - all extensions allowed except forbidden
		if len(s.allowedExtensions) > 0 {
			// not found in allowed
			if _, ok := s.allowedExtensions[ext]; !ok {
				next.ServeHTTP(w, r)
				return
			}

			// file extension allowed
		}

		// ok, file is not in the forbidden list
		// Stat it and get file info
		f, err := s.root.Open(fPath)
		if err != nil {
			// else no such file, show error in logs only in debug mode
			s.log.Debug("no such file or directory", "error", err)
			// pass request to the worker
			next.ServeHTTP(w, r)
			return
		}

		// at high confidence there is should not be an error
		// because we stat-ed the path previously and know, that that is file (not a dir), and it exists
		finfo, err := f.Stat()
		if err != nil {
			// else no such file, show error in logs only in debug mode
			s.log.Debug("no such file or directory", "error", err)
			// pass request to the worker
			next.ServeHTTP(w, r)
			return
		}

		defer func() {
			err = f.Close()
			if err != nil {
				s.log.Error("file close error", "error", err)
			}
		}()

		// if provided path to the dir, do not serve the dir, but pass the request to the worker
		if finfo.IsDir() {
			s.log.Debug("possible path to dir provided")
			// pass request to the worker
			next.ServeHTTP(w, r)
			return
		}

		// set etag
		if s.cfg.Static.CalculateEtag {
			SetEtag(s.cfg.Static.Weak, f, finfo.Name(), w)
		}

		// we passed all checks - serve the file
		http.ServeContent(w, r, finfo.Name(), finfo.ModTime(), f)
	})
}
