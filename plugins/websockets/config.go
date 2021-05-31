package websockets

import (
	"time"

	"github.com/spiral/roadrunner/v2/pkg/pool"
)

/*
websockets:
  # pubsubs should implement PubSub interface to be collected via endure.Collects

  pubsubs:["redis", "amqp", "memory"]
  # path used as websockets path
  path: "/ws"
*/

// Config represents configuration for the ws plugin
type Config struct {
	// http path for the websocket
	Path string `mapstructure:"path"`
	// ["redis", "amqp", "memory"]
	PubSubs    []string `mapstructure:"pubsubs"`
	Middleware []string `mapstructure:"middleware"`

	Pool *pool.Config `mapstructure:"pool"`
}

// InitDefault initialize default values for the ws config
func (c *Config) InitDefault() {
	if c.Path == "" {
		c.Path = "/ws"
	}
	if len(c.PubSubs) == 0 {
		// memory used by default
		c.PubSubs = append(c.PubSubs, "memory")
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
		if c.Pool.Supervisor == nil {
			return
		}
		c.Pool.Supervisor.InitDefaults()
	}
}
