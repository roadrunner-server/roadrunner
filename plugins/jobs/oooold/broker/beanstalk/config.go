package beanstalk

import (
	"fmt"
	"github.com/spiral/roadrunner/service"
	"strings"
	"time"
)

// Config defines beanstalk broker configuration.
type Config struct {
	// Addr of beanstalk server.
	Addr string

	// Timeout to allocate the connection. Default 10 seconds.
	Timeout int
}

// Hydrate config values.
func (c *Config) Hydrate(cfg service.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	if c.Addr == "" {
		return fmt.Errorf("beanstalk address is missing")
	}

	return nil
}

// TimeoutDuration returns number of seconds allowed to allocate the connection.
func (c *Config) TimeoutDuration() time.Duration {
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 10
	}

	return time.Duration(timeout) * time.Second
}

// size creates new rpc socket Listener.
func (c *Config) newConn() (*conn, error) {
	dsn := strings.Split(c.Addr, "://")
	if len(dsn) != 2 {
		return nil, fmt.Errorf("invalid socket DSN (tcp://localhost:11300, unix://beanstalk.sock)")
	}

	return newConn(dsn[0], dsn[1], c.TimeoutDuration())
}
