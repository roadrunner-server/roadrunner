package roadrunner

import (
	"errors"
	"net"
	"strings"
	"time"
	"os/exec"
)

const (
	FactoryPipes  = iota
	FactorySocket
)

// Server config combines factory, pool and cmd configurations.
type ServerConfig struct {
	// Command includes command strings with all the parameters, example: "php worker.php pipes". This config section
	//	// must not change on re-configuration.
	Command string

	// User specifies what user to run command under, for Unix systems only. Support both UID and name options. Keep
	// empty to use current user.This config section must not change on re-configuration.
	User string

	// Group specifies what group to run command under, for Unix systems only. Support GID or name options. Keep empty
	// to use current user.This config section must not change on re-configuration.
	Group string

	// Relay defines connection method and factory to be used to connect to workers:
	// "pipes", "tcp://:6001", "unix://rr.sock"
	// This config section must not change on re-configuration.
	Relay string

	// FactoryTimeout defines for how long socket factory will be waiting for worker connection. This config section
	// must not change on re-configuration.
	FactoryTimeout time.Duration

	// Pool defines worker pool configuration, number of workers, timeouts and etc. This config section might change
	// while server is running.
	Pool *Config
}

func (f *ServerConfig) makeCommand() (func() *exec.Cmd, error) {
	return nil, nil
}

// makeFactory creates and connects new factory instance based on given parameters.
func (f *ServerConfig) makeFactory() (Factory, error) {
	if f.Relay == "pipes" || f.Relay == "pipe" {
		return NewPipeFactory(), nil
	}

	dsn := strings.Split(f.Relay, "://")
	if len(dsn) != 2 {
		return nil, errors.New("invalid relay DSN (pipes, tcp://:6001, unix://rr.sock)")
	}

	ln, err := net.Listen(dsn[0], dsn[1])
	if err != nil {
		return nil, nil
	}

	return NewSocketFactory(ln, time.Second*f.FactoryTimeout), nil
}
