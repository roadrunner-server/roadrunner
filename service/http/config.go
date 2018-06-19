package http

import (
	"errors"
	"github.com/spiral/roadrunner"
	"strings"
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

// Valid validates the configuration.
func (cfg *Config) Valid() error {
	if cfg.Uploads == nil {
		return errors.New("mailformed uploads config")
	}

	if cfg.Workers == nil {
		return errors.New("mailformed workers config")
	}

	if cfg.Workers.Pool == nil {
		return errors.New("mailformed workers config (pool config is missing)")
	}

	if err := cfg.Workers.Pool.Valid(); err != nil {
		return err
	}

	if !strings.Contains(cfg.Address, ":") {
		return errors.New("mailformed server address")
	}

	return nil
}
