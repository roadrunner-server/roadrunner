package service

import (
	"reflect"
	"fmt"
)

const initMethod = "Init"

var noConfig = fmt.Errorf("no config has been provided")

// calls Init method with automatically resolved arguments.
func initService(s interface{}, cfg Config, c *container) (bool, error) {
	r := reflect.TypeOf(s)

	m, ok := r.MethodByName(initMethod)
	if !ok {
		// no Init method is presented, assuming service does not need
		// initialization.
		return false, nil
	}

	if err := verifySignature(m); err != nil {
		return false, err
	}

	// hydrating
	values, err := injectValues(m, s, cfg, c)
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

// injectValues returns slice of call arguments for service Init method.
func injectValues(m reflect.Method, s interface{}, cfg Config, c *container) (values []reflect.Value, err error) {
	for i := 0; i < m.Type.NumIn(); i ++ {
		v := m.Type.In(i)

		switch {
		case v.ConvertibleTo(reflect.ValueOf(s).Type()): // service itself
			values = append(values, reflect.ValueOf(s))

		case v.Implements(reflect.TypeOf((*HydrateConfig)(nil)).Elem()): // automatically configured config
			if cfg == nil {
				// todo: generic value
				return nil, noConfig
			}

			sc := reflect.New(v.Elem())
			if err := sc.Interface().(HydrateConfig).Hydrate(cfg); err != nil {
				return nil, err
			}

			values = append(values, sc)

		case v.Implements(reflect.TypeOf((*Config)(nil)).Elem()): // config section
			if cfg == nil {
				// todo: generic value
				return nil, noConfig
			}
			values = append(values, reflect.ValueOf(cfg))

		case v.Implements(reflect.TypeOf((*Container)(nil)).Elem()): // container
			values = append(values, reflect.ValueOf(c))

		default:
			found := false

			// looking for the service candidate
			for _, e := range c.services {
				if v.ConvertibleTo(reflect.ValueOf(e.svc).Type()) {
					found = true
					values = append(values, reflect.ValueOf(e.svc))
					break
				}
			}

			if !found {
				// placeholder (make sure to check inside the method)
				values = append(values, reflect.New(v).Elem())
			}
		}
	}

	return
}

// verifySignature checks if Init method has valid signature
func verifySignature(m reflect.Method) error {
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
