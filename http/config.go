package http

import (
	"github.com/spiral/roadrunner"
)

// Configures RoadRunner HTTP server.
type Config struct {
	// Enable enables http svc.
	Enable bool

	// Address and port to handle as http server.
	Address string

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
