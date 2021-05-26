package websockets

import "time"

/*
websockets:
  # pubsubs should implement PubSub interface to be collected via endure.Collects
  # also, they should implement RPC methods to publish data into them
  # pubsubs might use general config section or its own

  pubsubs:["redis", "amqp", "memory"]

  # sample of the own config section for the redis pubsub driver
  redis:
	address:
		- localhost:1111
  .... the rest


  # path used as websockets path
  path: "/ws"
*/

// Config represents configuration for the ws plugin
type Config struct {
	// http path for the websocket
	Path string `mapstructure:"path"`
	// ["redis", "amqp", "memory"]
	PubSubs    []string     `mapstructure:"pubsubs"`
	Middleware []string     `mapstructure:"middleware"`
	Redis      *RedisConfig `mapstructure:"redis"`
}

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

// InitDefault initialize default values for the ws config
func (c *Config) InitDefault() {
	if c.Path == "" {
		c.Path = "/ws"
	}
	if len(c.PubSubs) == 0 {
		// memory used by default
		c.PubSubs = append(c.PubSubs, "memory")
	}
}
