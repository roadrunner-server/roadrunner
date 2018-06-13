package static

import (
	"net/http"
	"path"
	"strings"
	rrttp "github.com/spiral/roadrunner/service/http"
	"github.com/spiral/roadrunner/service"
)

// ID contains default service name.
const ID = "static"

// Service serves static files. Potentially convert into middleware?
type Service struct {
	// server configuration (location, forbidden files and etc)
	cfg *Config

	// root is initiated http directory
	root http.Dir
}

// Init must return configure service and return true if service hasStatus enabled. Must return error in case of
// misconfiguration. Services must not be used without proper configuration pushed first.
func (s *Service) Init(cfg service.Config, c service.Container) (enabled bool, err error) {
	config := &Config{}
	if err := cfg.Unmarshal(config); err != nil {
		return false, err
	}

	if !config.Enable {
		return false, nil
	}

	if err := config.Valid(); err != nil {
		return false, err
	}

	s.cfg = config
	s.root = http.Dir(s.cfg.Dir)

	// registering as middleware
	if h, ok := c.Get(rrttp.ID); ok >= service.StatusConfigured {
		if h, ok := h.(*rrttp.Service); ok {
			h.AddMiddleware(s.middleware)
		}
	}

	return true, nil
}

// Serve serves the service.
func (s *Service) Serve() error { return nil }

// Stop stops the service.
func (s *Service) Stop() {}

// middleware must return true if request/response pair is handled withing the middleware.
func (s *Service) middleware(w http.ResponseWriter, r *http.Request) bool {
	fPath := r.URL.Path

	if !strings.HasPrefix(fPath, "/") {
		fPath = "/" + fPath
	}
	fPath = path.Clean(fPath)

	if s.cfg.Forbids(fPath) {
		return false
	}

	f, err := s.root.Open(fPath)
	if err != nil {
		return false
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		return false
	}

	// do not middleware directories
	if d.IsDir() {
		return false
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	return true
}
