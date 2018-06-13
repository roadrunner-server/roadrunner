package roadrunner

import (
	"errors"
	"net"
	"strings"
	"time"
	"os/exec"
)

// Server config combines factory, pool and cmd configurations.
type ServerConfig struct {
	// Command includes command strings with all the parameters, example: "php worker.php pipes".
	Command string

	// Relay defines connection method and factory to be used to connect to workers:
	// "pipes", "tcp://:6001", "unix://rr.sock"
	// This config section must not change on re-configuration.
	Relay string

	// RelayTimeout defines for how long socket factory will be waiting for worker connection. This config section
	// must not change on re-configuration.
	RelayTimeout time.Duration

	// Pool defines worker pool configuration, number of workers, timeouts and etc. This config section might change
	// while server is running.
	Pool *Config
}

// Differs returns true if configuration has changed but ignores pool or cmd changes.
func (cfg *ServerConfig) Differs(new *ServerConfig) bool {
	return cfg.Relay != new.Relay || cfg.RelayTimeout != new.RelayTimeout
}

// makeCommands returns new command provider based on configured options.
func (cfg *ServerConfig) makeCommand() func() *exec.Cmd {
	var cmd = strings.Split(cfg.Command, " ")
	return func() *exec.Cmd {
		return exec.Command(cmd[0], cmd[1:]...)
	}
}

// makeFactory creates and connects new factory instance based on given parameters.
func (cfg *ServerConfig) makeFactory() (Factory, error) {
	if cfg.Relay == "pipes" || cfg.Relay == "pipe" {
		return NewPipeFactory(), nil
	}

	dsn := strings.Split(cfg.Relay, "://")
	if len(dsn) != 2 {
		return nil, errors.New("invalid relay DSN (pipes, tcp://:6001, unix://rr.sock)")
	}

	ln, err := net.Listen(dsn[0], dsn[1])
	if err != nil {
		return nil, err
	}

	return NewSocketFactory(ln, cfg.RelayTimeout), nil
}
