package rpc

import (
	"errors"
	"net"
	"strings"

	"github.com/spiral/roadrunner/v2/utils"
)

// Config defines RPC service config.
type Config struct {
	// Listen string
	Listen string
}

// InitDefaults allows to init blank config with pre-defined set of default values.
func (c *Config) InitDefaults() {
	if c.Listen == "" {
		c.Listen = "tcp://127.0.0.1:6001"
	}
}

// Valid returns nil if config is valid.
func (c *Config) Valid() error {
	if dsn := strings.Split(c.Listen, "://"); len(dsn) != 2 {
		return errors.New("invalid socket DSN (tcp://:6001, unix://file.sock)")
	}

	return nil
}

// Listener creates new rpc socket Listener.
func (c *Config) Listener() (net.Listener, error) {
	return utils.CreateListener(c.Listen)
}

// Dialer creates rpc socket Dialer.
func (c *Config) Dialer() (net.Conn, error) {
	dsn := strings.Split(c.Listen, "://")
	if len(dsn) != 2 {
		return nil, errors.New("invalid socket DSN (tcp://:6001, unix://file.sock)")
	}

	return net.Dial(dsn[0], dsn[1])
}
