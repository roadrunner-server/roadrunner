package service

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync"
)

// Container controls all internal RR services and provides plugin based system.
type Container interface {
	// Register add new service to the container under given name.
	Register(name string, service interface{})

	// Reconfigure configures all underlying services with given configuration.
	Init(cfg Config) error

	// Check if svc has been registered.
	Has(service string) bool

	// get returns svc instance by it's name or nil if svc not found. Method returns current service status
	// as second value.
	Get(service string) (svc interface{}, status int)

	// Serve all configured services. Non blocking.
	Serve() error

	// Close all active services.
	Stop()
}

// Service can serve. Service can provide Init method which must return (bool, error) signature and might accept
// other services and/or configs as dependency. Container can be requested as well. Config can be requested in a form
// of service.Config or pointer to service specific config struct (automatically unmarshalled), config argument must
// implement service.HydrateConfig.
type Service interface {
	// Serve serves.
	Serve() error

	// Stop stops the service.
	Stop()
}

// Config provides ability to slice configuration sections and unmarshal configuration data into
// given structure.
type Config interface {
	// get nested config section (sub-map), returns nil if section not found.
	Get(service string) Config

	// Unmarshal unmarshal config data into given struct.
	Unmarshal(out interface{}) error
}

// HydrateConfig provides ability to automatically hydrate config with values using
// service.Config as the source.
type HydrateConfig interface {
	// Hydrate must populate config values using given config source.
	// Must return error if config is not valid.
	Hydrate(cfg Config) error
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
func (c *container) Register(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = append(c.services, &entry{
		name:   name,
		svc:    service,
		status: StatusRegistered,
	})

	c.log.Debugf("[%s]: registered", name)
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

// get returns svc instance by it's name or nil if svc not found.
func (c *container) Get(target string) (svc interface{}, status int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, e := range c.services {
		if e.name == target {
			return e.svc, e.getStatus()
		}
	}

	return nil, StatusUndefined
}

// Init configures all underlying services with given configuration.
func (c *container) Init(cfg Config) error {
	for _, e := range c.services {
		if e.getStatus() >= StatusOK {
			return fmt.Errorf("service [%s] has already been configured", e.name)
		}

		// inject service dependencies (todo: move to container)
		if ok, err := initService(e.svc, cfg.Get(e.name), c); err != nil {
			if err == noConfig {
				c.log.Warningf("[%s]: no config has been provided", e.name)

				// unable to meet dependency requirements, skippingF
				continue
			}

			return errors.Wrap(err, fmt.Sprintf("[%s]", e.name))
		} else if ok {
			e.setStatus(StatusOK)
			c.log.Debugf("[%s]: initiated", e.name)
		} else {
			c.log.Debugf("[%s]: disabled", e.name)
		}
	}

	return nil
}

//todo: refactor ????
// Serve all configured services. Non blocking.
func (c *container) Serve() error {
	var (
		numServing int
		done       = make(chan interface{}, len(c.services))
	)

	for _, e := range c.services {
		if e.hasStatus(StatusOK) && e.canServe() {
			numServing++
		} else {
			continue
		}

		c.log.Debugf("[%s]: service started", e.name)
		go func(e *entry) {
			e.setStatus(StatusServing)
			defer e.setStatus(StatusStopped)

			if err := e.svc.(Service).Serve(); err != nil {
				c.log.Errorf("[%s]: %s", e.name, err)
				done <- errors.Wrap(err, fmt.Sprintf("[%s]", e.name))
			} else {
				done <- nil
			}
		}(e)
	}

	for i := 0; i < numServing; i++ {
		result := <-done

		if result == nil {
			// no errors
			continue
		}

		// found an error in one of the services, stopping the rest of running services.
		if err := result.(error); err != nil {
			c.Stop()
			return err
		}
	}

	return nil
}

// Stop sends stop command to all running services.
func (c *container) Stop() {
	c.log.Debugf("received stop command")
	for _, e := range c.services {
		if e.hasStatus(StatusServing) {
			e.svc.(Service).Stop()
			e.setStatus(StatusStopped)

			c.log.Debugf("[%s]: stopped", e.name)
		}
	}
}
