package http

import (
	"github.com/spiral/roadrunner"
	"fmt"
)

// Configures RoadRunner HTTP server.
type Config struct {
	// Enable enables http service.
	Enable bool

	// Host and port to handle as http server.
	Host, Port string

	// MaxRequest specified max size for payload body in bytes, set 0 to unlimited.
	MaxRequest int64

	// Uploads configures uploads configuration.
	Uploads *UploadsConfig

	// Workers configures roadrunner server and worker pool.
	Workers *roadrunner.ServerConfig
}

// Valid validates the configuration.
func (cfg *Config) Valid() error {
	return nil
}

// httpAddr returns prepared http listen address.
func (cfg *Config) httpAddr() string {
	return fmt.Sprintf("%s:%v", cfg.Host, cfg.Port)
}
