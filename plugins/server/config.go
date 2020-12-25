package server

import (
	"time"
)

// Config config combines factory, pool and cmd configurations.
type Config struct {
	// Command to run as application.
	Command string

	// User to run application under.
	User string

	// Group to run application under.
	Group string

	// Env represents application environment.
	Env Env

	// Listen defines connection method and factory to be used to connect to workers:
	// "pipes", "tcp://:6001", "unix://rr.sock"
	// This config section must not change on re-configuration.
	Relay string

	// RelayTimeout defines for how long socket factory will be waiting for worker connection. This config section
	// must not change on re-configuration. Defaults to 60s.
	RelayTimeout time.Duration
}

func (cfg *Config) InitDefaults() {
	if cfg.Relay == "" {
		cfg.Relay = "pipes"
	}

	if cfg.RelayTimeout == 0 {
		cfg.RelayTimeout = time.Second * 60
	}
}
