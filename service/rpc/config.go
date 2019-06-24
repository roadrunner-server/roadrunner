package rpc

import (
	"errors"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/util"
	"net"
	"strings"
)

// Config defines RPC service config.
type Config struct {
	// Indicates if RPC connection is enabled.
	Enable bool

	// Listen string
	Listen string
}

// Hydrate must populate Config values using given Config source. Must return error if Config is not valid.
func (c *Config) Hydrate(cfg service.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// InitDefaults allows to init blank config with pre-defined set of default values.
func (c *Config) InitDefaults() error {
	c.Enable = true
	c.Listen = "tcp://127.0.0.1:6001"

	return nil
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
	return util.CreateListener(c.Listen)
}

// Dialer creates rpc socket Dialer.
func (c *Config) Dialer() (net.Conn, error) {
	dsn := strings.Split(c.Listen, "://")
	if len(dsn) != 2 {
		return nil, errors.New("invalid socket DSN (tcp://:6001, unix://file.sock)")
	}

	return net.Dial(dsn[0], dsn[1])
}
