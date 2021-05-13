package server

import (
	"time"
)

// Config All config (.rr.yaml)
// For other section use pointer to distinguish between `empty` and `not present`
type Config struct {
	// Server config section
	Server struct {
		// Command to run as application.
		Command string `mapstructure:"command"`
		// User to run application under.
		User string `mapstructure:"user"`
		// Group to run application under.
		Group string `mapstructure:"group"`
		// Env represents application environment.
		Env Env `mapstructure:"env"`
		// Relay defines connection method and factory to be used to connect to workers:
		// "pipes", "tcp://:6001", "unix://rr.sock"
		// This config section must not change on re-configuration.
		Relay string `mapstructure:"relay"`
		// RelayTimeout defines for how long socket factory will be waiting for worker connection. This config section
		// must not change on re-configuration. Defaults to 60s.
		RelayTimeout time.Duration `mapstructure:"relay_timeout"`
	} `mapstructure:"server"`

	// we just need to know if the section exist, we don't need to read config from it
	RPC *struct {
		Listen string `mapstructure:"listen"`
	} `mapstructure:"rpc"`
	Logs *struct {
	} `mapstructure:"logs"`
	HTTP *struct {
	} `mapstructure:"http"`
	Redis *struct {
	} `mapstructure:"redis"`
	Boltdb *struct {
	} `mapstructure:"boltdb"`
	Memcached *struct {
	} `mapstructure:"memcached"`
	Memory *struct {
	} `mapstructure:"memory"`
	Metrics *struct {
	} `mapstructure:"metrics"`
	Reload *struct {
	} `mapstructure:"reload"`
}

// InitDefaults for the server config
func (cfg *Config) InitDefaults() {
	if cfg.Server.Relay == "" {
		cfg.Server.Relay = "pipes"
	}

	if cfg.Server.RelayTimeout == 0 {
		cfg.Server.RelayTimeout = time.Second * 60
	}
}
