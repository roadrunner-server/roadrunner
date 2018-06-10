package service

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync"
	"github.com/fatih/color"
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
func (c *container) Register(name string, service Service) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = append(c.services, &entry{
		name:   name,
		svc:    service,
		status: StatusRegistered,
	})

	c.log.Debugf("%s: registered", color.YellowString(name))
}

// Check hasStatus svc has been registered.
func (c *container) Has(target string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, e := range c.services {
		if e.name == target {
			return true
		}
	}

	return false
}

// Get returns svc instance by it's name or nil if svc not found.
func (c *container) Get(target string) (svc Service, status int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, e := range c.services {
		if e.name == target {
			return e.svc, e.getStatus()
		}
	}

	return nil, StatusUndefined
}

// Configure configures all underlying services with given configuration.
func (c *container) Configure(cfg Config) error {
	for _, e := range c.services {
		if e.getStatus() >= StatusConfigured {
			return fmt.Errorf("service %s has already been configured", color.RedString(e.name))
		}

		segment := cfg.Get(e.name)
		if segment == nil {
			c.log.Debugf("%s: no config has been provided", color.YellowString(e.name))
			continue
		}

		ok, err := e.svc.Configure(segment, c)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("%s", color.RedString(e.name)))
		} else if ok {
			e.setStatus(StatusConfigured)
		}
	}

	return nil
}

// Serve all configured services. Non blocking.
func (c *container) Serve() error {
	var (
		numServing int
		done       = make(chan interface{}, len(c.services))
	)
	defer close(done)

	for _, e := range c.services {
		if e.hasStatus(StatusConfigured) {
			numServing ++
		} else {
			continue
		}

		c.log.Debugf("%s: started", color.GreenString(e.name))
		go func(e *entry) {
			e.setStatus(StatusServing)
			defer e.setStatus(StatusStopped)

			if err := e.svc.Serve(); err != nil {
				done <- errors.Wrap(err, fmt.Sprintf("%s", color.RedString(e.name)))
			}
		}(e)
	}

	for i := 0; i < numServing; i++ {
		result := <-done

		// found an error in one of the services, stopping the rest of running services.
		if err, ok := result.(error); ok {
			c.Stop()
			return err
		}
	}

	return nil
}

// Stop sends stop command to all running services.
func (c *container) Stop() {
	for _, e := range c.services {
		if e.hasStatus(StatusServing) {
			e.svc.Stop()
			e.setStatus(StatusStopped)
			c.log.Debugf("%s: stopped", color.GreenString(e.name))
		}
	}
}
