package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	rrhttp "github.com/spiral/roadrunner/service/http"
)

// ID declares the public service name
const ID = "health"

// Service to serve an endpoint for checking the health of the worker pool
type Service struct {
	cfg         *Config
	mu          sync.Mutex
	http        *http.Server
	httpService *rrhttp.Service
}

// Init health service
func (s *Service) Init(cfg *Config, r *rrhttp.Service) (bool, error) {
	// Ensure the httpService is set
	if r == nil {
		return false, nil
	}

	s.cfg = cfg
	s.httpService = r
	return true, nil
}

// Serve the health endpoint
func (s *Service) Serve() error {
	// Configure and start the http server
	s.mu.Lock()
	s.http = &http.Server{Addr: s.cfg.Address, Handler: s}
	s.mu.Unlock()
	return s.http.ListenAndServe()
}

// Stop the health endpoint
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.http != nil {
		// gracefully stop the server
		go func() {
			err := s.http.Shutdown(context.Background())
			if err != nil {
				// TODO how to log error here?
				fmt.Println(fmt.Errorf("error shutting down the server: error %v", err))
			}
		}()
	}
}

// ServeHTTP returns the health of the pool of workers
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	if !s.isHealthy() {
		status = http.StatusInternalServerError
	}
	w.WriteHeader(status)
}

// isHealthy checks the server, pool and ensures at least one worker is active
func (s *Service) isHealthy() bool {
	httpService := s.httpService
	if httpService == nil {
		return false
	}

	server := httpService.Server()
	if server == nil {
		return false
	}

	pool := server.Pool()
	if pool == nil {
		return false
	}

	// Ensure at least one worker is active
	for _, w := range pool.Workers() {
		if w.State().IsActive() {
			return true
		}
	}

	return false
}
