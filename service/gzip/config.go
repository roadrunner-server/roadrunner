package gzip

import (
	"github.com/spiral/roadrunner/service"
)

// Config describes file location and controls access to them.
type Config struct {
	Enable bool
}

// Hydrate must populate Config values using given Config source. Must return error if Config is not valid.
func (c *Config) Hydrate(cfg service.Config) error {
	return cfg.Unmarshal(c)
}

// InitDefaults sets missing values to their default values.
func (c *Config) InitDefaults() error {
	c.Enable = true

	return nil
}
