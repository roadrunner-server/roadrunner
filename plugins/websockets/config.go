package websockets

import (
	"strings"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pool"
)

/*
websockets:
  broker: default
  allowed_origin: "*"
  path: "/ws"
*/

// Config represents configuration for the ws plugin
type Config struct {
	// http path for the websocket
	Path          string `mapstructure:"path"`
	AllowedOrigin string `mapstructure:"allowed_origin"`
	Broker        string `mapstructure:"broker"`

	// wildcard origin
	allowedWOrigins []wildcard
	allowedOrigins  []string
	allowedAll      bool

	// Pool with the workers for the websockets
	Pool *pool.Config `mapstructure:"pool"`
}

// InitDefault initialize default values for the ws config
func (c *Config) InitDefault() error {
	if c.Path == "" {
		c.Path = "/ws"
	}

	// broker is mandatory
	if c.Broker == "" {
		return errors.Str("broker key should be specified")
	}

	if c.Pool == nil {
		c.Pool = &pool.Config{}
		if c.Pool.NumWorkers == 0 {
			// 2 workers by default
			c.Pool.NumWorkers = 2
		}

		if c.Pool.AllocateTimeout == 0 {
			c.Pool.AllocateTimeout = time.Minute
		}

		if c.Pool.DestroyTimeout == 0 {
			c.Pool.DestroyTimeout = time.Minute
		}
		if c.Pool.Supervisor != nil {
			c.Pool.Supervisor.InitDefaults()
		}
	}

	if c.AllowedOrigin == "" {
		c.AllowedOrigin = "*"
	}

	// Normalize
	origin := strings.ToLower(c.AllowedOrigin)
	if origin == "*" {
		// If "*" is present in the list, turn the whole list into a match all
		c.allowedAll = true
		return nil
	} else if i := strings.IndexByte(origin, '*'); i >= 0 {
		// Split the origin in two: start and end string without the *
		w := wildcard{origin[0:i], origin[i+1:]}
		c.allowedWOrigins = append(c.allowedWOrigins, w)
	} else {
		c.allowedOrigins = append(c.allowedOrigins, origin)
	}

	return nil
}
