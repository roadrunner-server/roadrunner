package roadrunner

import (
	"errors"
	"net"
	"strings"
	"time"
)

const (
	FactoryPipes = iota
	FactorySocket
)

type ServerConfig struct {
	// Relay defines connection method and factory to be used to connect to workers:
	// "pipes", "tcp://:6001", "unix://rr.sock"
	// This config section must not change on re-configuration.
	Relay string

	// FactoryTimeout defines for how long socket factory will be waiting for worker connection. This config section
	// must not change on re-configuration.
	FactoryTimeout time.Duration

	// Pool defines worker pool configuration, number of workers, timeouts and etc. This config section might change
	// while server is running.
	Pool Config
}

// buildFactory creates and connects new factory instance based on given parameters.
func (f *ServerConfig) buildFactory() (Factory, error) {
	if f.Relay == "pipes" {
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
