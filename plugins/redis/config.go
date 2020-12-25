package redis

import "time"

type Config struct {
	Addrs            []string      `yaml:"addrs"`
	DB               int           `yaml:"db"`
	Username         string        `yaml:"username"`
	Password         string        `yaml:"password"`
	MasterName       string        `yaml:"master_name"`
	SentinelPassword string        `yaml:"sentinel_password"`
	RouteByLatency   bool          `yaml:"route_by_latency"`
	RouteRandomly    bool          `yaml:"route_randomly"`
	MaxRetries       int           `yaml:"max_retries"`
	DialTimeout      time.Duration `yaml:"dial_timeout"`
	MinRetryBackoff  time.Duration `yaml:"min_retry_backoff"`
	MaxRetryBackoff  time.Duration `yaml:"max_retry_backoff"`
	PoolSize         int           `yaml:"pool_size"`
	MinIdleConns     int           `yaml:"min_idle_conns"`
	MaxConnAge       time.Duration `yaml:"max_conn_age"`
	ReadTimeout      time.Duration `yaml:"read_timeout"`
	WriteTimeout     time.Duration `yaml:"write_timeout"`
	PoolTimeout      time.Duration `yaml:"pool_timeout"`
	IdleTimeout      time.Duration `yaml:"idle_timeout"`
	IdleCheckFreq    time.Duration `yaml:"idle_check_freq"`
	ReadOnly         bool          `yaml:"read_only"`
}

// InitDefaults initializing fill config with default values
func (s *Config) InitDefaults() {
	s.Addrs = []string{"localhost:6379"} // default addr is pointing to local storage
}
