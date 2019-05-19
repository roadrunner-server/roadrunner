package service

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"reflect"
	"sync"
)

var errNoConfig = fmt.Errorf("no config has been provided")

// InitMethod contains name of the method to be automatically invoked while service initialization. Must return
// (bool, error). Container can be requested as well. Config can be requested in a form
// of service.Config or pointer to service specific config struct (automatically unmarshalled), config argument must
// implement service.HydrateConfig.
const InitMethod = "Init"

// Service can serve. Services can provide Init method which must return (bool, error) signature and might accept
// other services and/or configs as dependency.
type Service interface {
	// Serve serves.
	Serve() error

	// Detach stops the service.
	Stop()
}

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

// DefaultsConfig declares ability to be initated without config data provided.
type DefaultsConfig interface {
	// InitDefaults allows to init blank config with pre-defined set of default values.
	InitDefaults() error
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
		status: StatusInactive,
	})
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

		// inject service dependencies
		if ok, err := c.initService(e.svc, cfg.Get(e.name)); err != nil {
			// soft error (skipping)
			if err == errNoConfig {
				c.log.Debugf("[%s]: disabled", e.name)
				continue
			}

			return errors.Wrap(err, fmt.Sprintf("[%s]", e.name))
		} else if ok {
			e.setStatus(StatusOK)
		} else {
			c.log.Debugf("[%s]: disabled", e.name)
		}
	}

	return nil
}

// Serve all configured services. Non blocking.
func (c *container) Serve() error {
	var (
		numServing = 0
		done       = make(chan interface{}, len(c.services))
	)

	for _, e := range c.services {
		if e.hasStatus(StatusOK) && e.canServe() {
			numServing++
		} else {
			continue
		}

		c.log.Debugf("[%s]: started", e.name)
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

// Detach sends stop command to all running services.
func (c *container) Stop() {
	for _, e := range c.services {
		if e.hasStatus(StatusServing) {
			e.setStatus(StatusStopping)
			e.svc.(Service).Stop()
			e.setStatus(StatusStopped)

			c.log.Debugf("[%s]: stopped", e.name)
		}
	}
}

// calls Init method with automatically resolved arguments.
func (c *container) initService(s interface{}, segment Config) (bool, error) {
	r := reflect.TypeOf(s)

	m, ok := r.MethodByName(InitMethod)
	if !ok {
		// no Init method is presented, assuming service does not need initialization.
		return true, nil
	}

	if err := c.verifySignature(m); err != nil {
		return false, err
	}

	// hydrating
	values, err := c.resolveValues(s, m, segment)
	if err != nil {
		return false, err
	}

	// initiating service
	out := m.Func.Call(values)

	if out[1].IsNil() {
		return out[0].Bool(), nil
	}

	return out[0].Bool(), out[1].Interface().(error)
}

// resolveValues returns slice of call arguments for service Init method.
func (c *container) resolveValues(s interface{}, m reflect.Method, cfg Config) (values []reflect.Value, err error) {
	for i := 0; i < m.Type.NumIn(); i++ {
		v := m.Type.In(i)

		switch {
		case v.ConvertibleTo(reflect.ValueOf(s).Type()): // service itself
			values = append(values, reflect.ValueOf(s))

		case v.Implements(reflect.TypeOf((*Container)(nil)).Elem()): // container
			values = append(values, reflect.ValueOf(c))

		case v.Implements(reflect.TypeOf((*logrus.StdLogger)(nil)).Elem()),
			v.Implements(reflect.TypeOf((*logrus.FieldLogger)(nil)).Elem()),
			v.ConvertibleTo(reflect.ValueOf(c.log).Type()): // logger
			values = append(values, reflect.ValueOf(c.log))

		case v.Implements(reflect.TypeOf((*HydrateConfig)(nil)).Elem()): // injectable config
			sc := reflect.New(v.Elem())

			if dsc, ok := sc.Interface().(DefaultsConfig); ok {
				dsc.InitDefaults()
				if cfg == nil {
					values = append(values, sc)
					continue
				}

			} else if cfg == nil {
				return nil, errNoConfig
			}

			if err := sc.Interface().(HydrateConfig).Hydrate(cfg); err != nil {
				return nil, err
			}

			values = append(values, sc)

		case v.Implements(reflect.TypeOf((*Config)(nil)).Elem()): // generic config section
			if cfg == nil {
				return nil, errNoConfig
			}

			values = append(values, reflect.ValueOf(cfg))

		default: // dependency on other service (resolution to nil if service can't be found)
			value, err := c.resolveValue(v)
			if err != nil {
				return nil, err
			}

			values = append(values, value)
		}
	}

	return
}

// verifySignature checks if Init method has valid signature
func (c *container) verifySignature(m reflect.Method) error {
	if m.Type.NumOut() != 2 {
		return fmt.Errorf("method Init must have exact 2 return values")
	}

	if m.Type.Out(0).Kind() != reflect.Bool {
		return fmt.Errorf("first return value of Init method must be bool type")
	}

	if !m.Type.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return fmt.Errorf("second return value of Init method value must be error type")
	}

	return nil
}

func (c *container) resolveValue(v reflect.Type) (reflect.Value, error) {
	value := reflect.Value{}
	for _, e := range c.services {
		if !e.hasStatus(StatusOK) {
			continue
		}

		if v.Kind() == reflect.Interface && reflect.TypeOf(e.svc).Implements(v) {
			if value.IsValid() {
				return value, fmt.Errorf("disambiguous dependency `%s`", v)
			}

			value = reflect.ValueOf(e.svc)
		}

		if v.ConvertibleTo(reflect.ValueOf(e.svc).Type()) {
			if value.IsValid() {
				return value, fmt.Errorf("disambiguous dependency `%s`", v)
			}

			value = reflect.ValueOf(e.svc)
		}
	}

	if !value.IsValid() {
		// placeholder (make sure to check inside the method)
		value = reflect.New(v).Elem()
	}

	return value, nil
}
