package health

import (
	"errors"
	"strings"

	"github.com/spiral/roadrunner/service"
)

// Config configures the health service
type Config struct {
	// Address to listen on
	Address string
}

// Hydrate the config
func (c *Config) Hydrate(cfg service.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}
	return c.Valid()
}

// Valid validates the configuration.
func (c *Config) Valid() error {
	// Validate the address
	if c.Address != "" && !strings.Contains(c.Address, ":") {
		return errors.New("malformed http server address")
	}

	return nil
}
