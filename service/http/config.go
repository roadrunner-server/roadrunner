package http

import (
	"errors"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service"
	"strings"
	"time"
)

// Config configures RoadRunner HTTP server.
type Config struct {
	// Enable enables http svc.
	Enable bool

	// Address and port to handle as http server.
	Address string

	// MaxRequest specified max size for payload body in megabytes, set 0 to unlimited.
	MaxRequest int64

	// Uploads configures uploads configuration.
	Uploads *UploadsConfig

	// Workers configures roadrunner server and worker pool.
	Workers *roadrunner.ServerConfig
}

// Hydrate must populate Config values using given Config source. Must return error if Config is not valid.
func (c *Config) Hydrate(cfg service.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	if c.Workers.Relay == "" {
		c.Workers.Relay = "pipes"
	}

	if err := c.Valid(); err != nil {
		return err
	}

	if c.Workers.RelayTimeout < time.Microsecond {
		c.Workers.RelayTimeout = time.Second * time.Duration(c.Workers.RelayTimeout.Nanoseconds())
	}

	if c.Workers.Pool.AllocateTimeout < time.Microsecond {
		c.Workers.Pool.AllocateTimeout = time.Second * time.Duration(c.Workers.Pool.AllocateTimeout.Nanoseconds())
	}

	if c.Workers.Pool.DestroyTimeout < time.Microsecond {
		c.Workers.Pool.DestroyTimeout = time.Second * time.Duration(c.Workers.Pool.DestroyTimeout.Nanoseconds())
	}

	return nil
}

// Valid validates the configuration.
func (c *Config) Valid() error {
	if c.Uploads == nil {
		return errors.New("mailformed uploads config")
	}

	if c.Workers == nil {
		return errors.New("mailformed workers config")
	}

	if c.Workers.Pool == nil {
		return errors.New("mailformed workers config (pool config is missing)")
	}

	if err := c.Workers.Pool.Valid(); err != nil {
		return err
	}

	if !strings.Contains(c.Address, ":") {
		return errors.New("mailformed server address")
	}

	return nil
}
