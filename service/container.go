package service

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync"
)

// Config provides ability to slice configuration sections and unmarshal configuration data into
// given structure.
type Config interface {
	// Get nested config section (sub-map), returns nil if section not found.
	Get(service string) Config

	// Unmarshal unmarshal config data into given struct.
	Unmarshal(out interface{}) error
}

// Container controls all internal RR services and provides plugin based system.
type Container interface {
	// Register add new service to the container under given name.
	Register(name string, service Service)

	// Reconfigure configures all underlying services with given configuration.
	Configure(cfg Config) error

	// Check if svc has been registered.
	Has(service string) bool

	// Get returns svc instance by it's name or nil if svc not found. Method returns current service status
	// as second value.
	Get(service string) (svc Service, status int)

	// Serve all configured services. Non blocking.
	Serve() error

	// Close all active services.
	Stop()
}

// svc provides high level functionality for road runner svc.
type Service interface {
	// Configure must return configure service and return true if service hasStatus enabled. Must return error in case of
	// misconfiguration. Services must not be used without proper configuration pushed first.
	Configure(cfg Config, reg Container) (enabled bool, err error)

	// Serve serves svc.
	Serve() error

	// Close setStopped svc svc.
	Stop()
}

type container struct {
	log      logrus.FieldLogger
	mu       sync.Mutex
	services []*entry
}

// NewContainer creates new service container.
func NewContainer(log logrus.FieldLogger) Container {
	return &container{
		log:      log,
		services: make([]*entry, 0),
	}
}

// Register add new service to the container under given name.
func (r *container) Register(name string, service Service) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.services = append(r.services, &entry{
		name:   name,
		svc:    service,
		status: StatusConfigured,
	})

	r.log.Debugf("%s.service: registered", name)
}

// Check hasStatus svc has been registered.
func (r *container) Has(target string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, e := range r.services {
		if e.name == target {
			return true
		}
	}

	return false
}

// Get returns svc instance by it's name or nil if svc not found.
func (r *container) Get(target string) (svc Service, status int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, e := range r.services {
		if e.name == target {
			return e.svc, e.getStatus()
		}
	}

	return nil, StatusUndefined
}

// Configure configures all underlying services with given configuration.
func (r *container) Configure(cfg Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, e := range r.services {
		if e.getStatus() >= StatusConfigured {
			return fmt.Errorf("service %s has already been configured", e.name)
		}

		segment := cfg.Get(e.name)
		if segment == nil {
			r.log.Debugf("%s.service: no config has been provided", e.name)
			continue
		}

		ok, err := e.svc.Configure(segment, r)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("%s.service", e.name))
		} else if ok {
			e.setStatus(StatusConfigured)
		}
	}

	return nil
}

// Serve all configured services. Non blocking.
func (r *container) Serve() error {
	var (
		numServing int
		done       = make(chan interface{}, len(r.services))
	)
	defer close(done)

	r.mu.Lock()
	for _, e := range r.services {
		if e.hasStatus(StatusConfigured) {
			numServing ++
		} else {
			continue
		}

		go func(s *entry) {
			s.setStatus(StatusServing)
			defer s.setStatus(StatusStopped)

			if err := s.svc.Serve(); err != nil {
				done <- err
			}
		}(e)
	}
	r.mu.Unlock()

	for i := 0; i < numServing; i++ {
		result := <-done

		// found an error in one of the services, stopping the rest of running services.
		if err, ok := result.(error); ok {
			r.Stop()
			return err
		}
	}

	return nil
}

// Stop sends stop command to all running services.
func (r *container) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, e := range r.services {
		if e.hasStatus(StatusServing) {
			e.svc.Stop()
		}
	}
}
