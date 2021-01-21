package redis

import "time"

type Config struct {
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

// InitDefaults initializing fill config with default values
func (s *Config) InitDefaults() {
	if s.Addrs == nil {
		s.Addrs = []string{"localhost:6379"} // default addr is pointing to local storage
	}
}
