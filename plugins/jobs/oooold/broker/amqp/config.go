package amqp

import (
	"fmt"
	"github.com/spiral/roadrunner/service"
	"time"
)

// Config defines sqs broker configuration.
type Config struct {
	// Addr of AMQP server (example: amqp://guest:guest@localhost:5672/).
	Addr string

	// Timeout to allocate the connection. Default 10 seconds.
	Timeout int
}

// Hydrate config values.
func (c *Config) Hydrate(cfg service.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	if c.Addr == "" {
		return fmt.Errorf("AMQP address is missing")
	}

	return nil
}

// TimeoutDuration returns number of seconds allowed to redial
func (c *Config) TimeoutDuration() time.Duration {
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 10
	}

	return time.Duration(timeout) * time.Second
}
