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

// Registry controls all internal RR services and provides plugin based system.
type Registry interface {
	// Register add new service to the registry under given name.
	Register(name string, service Service)

	// Reconfigure configures all underlying services with given configuration.
	Configure(cfg Config) error

	// Check is Service has been registered and configured.
	Has(service string) bool

	// Get returns Service instance by it's Name or nil if Service not found. Method must return only configured instance.
	Get(service string) Service

	// Serve all configured services. Non blocking.
	Serve() error

	// Close all active services.
	Stop() error
}

// Service provides high level functionality for road runner Service.
type Service interface {
	// WithConfig must return Service instance configured with the given environment. Must return error in case of
	// misconfiguration, might return nil as Service if Service is not enabled.
	WithConfig(cfg Config, reg Registry) (Service, error)

	// Serve serves Service.
	Serve() error

	// Close stop Service Service.
	Stop() error
}

type registry struct {
	log        logrus.FieldLogger
	mu         sync.Mutex
	candidates []*entry
	configured []*entry
}

// entry creates association between service instance and given name.
type entry struct {
	// Associated service name
	Name string

	// Associated service instance
	Service Service

	// Serving indicates that service is currently serving
	Serving bool
}

// NewRegistry creates new registry.
func NewRegistry(log logrus.FieldLogger) Registry {
	return &registry{
		log:        log,
		candidates: make([]*entry, 0),
	}
}

// Register add new service to the registry under given name.
func (r *registry) Register(name string, service Service) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.candidates = append(r.candidates, &entry{
		Name:    name,
		Service: service,
		Serving: false,
	})

	r.log.Debugf("%s.service: registered", name)
}

// Reconfigure configures all underlying services with given configuration.
func (r *registry) Configure(cfg Config) error {
	if r.configured != nil {
		return fmt.Errorf("service bus has been already configured")
	}

	r.configured = make([]*entry, 0)
	for _, e := range r.candidates {
		segment := cfg.Get(e.Name)
		if segment == nil {
			r.log.Debugf("%s.service: no config has been provided", e.Name)
			continue
		}

		s, err := e.Service.WithConfig(segment, r)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("%s.service", e.Name))
		}

		if s != nil {
			r.configured = append(r.configured, &entry{
				Name:    e.Name,
				Service: s,
				Serving: false,
			})
		}
	}

	return nil
}

// Check is Service has been registered.
func (r *registry) Has(service string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, e := range r.configured {
		if e.Name == service {
			return true
		}
	}

	return false
}

// Get returns Service instance by it's Name or nil if Service not found.
func (r *registry) Get(service string) Service {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, e := range r.configured {
		if e.Name == service {
			return e.Service
		}
	}

	return nil
}

// Serve all configured services. Non blocking.
func (r *registry) Serve() error {
	return nil
}

// Close all active services.
func (r *registry) Stop() error {
	return nil
}
