package env

import (
	"github.com/spiral/roadrunner/service"
)

// Config defines set of env values for RR workers.
type Config struct {
	// values to set as worker _ENV.
	Values map[string]string
}

// Hydrate must populate Config values using given Config source. Must return error if Config is not valid.
func (c *Config) Hydrate(cfg service.Config) error {
	return cfg.Unmarshal(&c.Values)
}

// InitDefaults allows to init blank config with pre-defined set of default values.
func (c *Config) InitDefaults() error {
	c.Values = make(map[string]string)
	return nil
}
