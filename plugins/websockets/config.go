package websockets

import (
	"strings"
	"time"

	"github.com/spiral/roadrunner/v2/pkg/pool"
)

/*
# GLOBAL
redis:
  addrs:
    - 'localhost:6379'

websockets:
  # pubsubs should implement PubSub interface to be collected via endure.Collects

  pubsubs:["redis", "amqp", "memory"]
  # OR local
  redis:
    addrs:
      - 'localhost:6379'

  # path used as websockets path
  path: "/ws"
*/

type RedisConfig struct {
	Addrs            []string      `mapstructure:"addrs"`
	DB               int           `mapstructure:"db"`
	Username         string        `mapstructure:"username"`
	Password         string        `mapstructure:"password"`
	MasterName       string        `mapstructure:"master_name"`
	SentinelPassword string        `mapstructure:"sentinel_password"`
	RouteByLatency   bool          `mapstructure:"route_by_latency"`
	RouteRandomly    bool          `mapstructure:"route_randomly"`
	MaxRetries       int           `mapstructure:"max_retries"`
	DialTimeout      time.Duration `mapstructure:"dial_timeout"`
	MinRetryBackoff  time.Duration `mapstructure:"min_retry_backoff"`
	MaxRetryBackoff  time.Duration `mapstructure:"max_retry_backoff"`
	PoolSize         int           `mapstructure:"pool_size"`
	MinIdleConns     int           `mapstructure:"min_idle_conns"`
	MaxConnAge       time.Duration `mapstructure:"max_conn_age"`
	ReadTimeout      time.Duration `mapstructure:"read_timeout"`
	WriteTimeout     time.Duration `mapstructure:"write_timeout"`
	PoolTimeout      time.Duration `mapstructure:"pool_timeout"`
	IdleTimeout      time.Duration `mapstructure:"idle_timeout"`
	IdleCheckFreq    time.Duration `mapstructure:"idle_check_freq"`
	ReadOnly         bool          `mapstructure:"read_only"`
}

// Config represents configuration for the ws plugin
type Config struct {
	// http path for the websocket
	Path string `mapstructure:"path"`
	// ["redis", "amqp", "memory"]
	PubSubs    []string `mapstructure:"pubsubs"`
	Middleware []string `mapstructure:"middleware"`

	AllowedOrigin string `mapstructure:"allowed_origin"`

	// wildcard origin
	allowedWOrigins []wildcard
	allowedOrigins  []string
	allowedAll      bool

	Redis *RedisConfig `mapstructure:"redis"`
	Pool  *pool.Config `mapstructure:"pool"`
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
		if c.Pool.Supervisor != nil {
			c.Pool.Supervisor.InitDefaults()
		}
	}

	if c.Redis != nil {
		if c.Redis.Addrs == nil {
			// append default
			c.Redis.Addrs = append(c.Redis.Addrs, "localhost:6379")
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
		return
	} else if i := strings.IndexByte(origin, '*'); i >= 0 {
		// Split the origin in two: start and end string without the *
		w := wildcard{origin[0:i], origin[i+1:]}
		c.allowedWOrigins = append(c.allowedWOrigins, w)
	} else {
		c.allowedOrigins = append(c.allowedOrigins, origin)
	}
}
