package rpc

import (
	"errors"
	"net"
	"strings"
	"syscall"
)

type config struct {
	// Indicates if RPC connection is enabled.
	Enable bool

	// AddListener string
	Listen string
}

// listener creates new rpc socket listener.
func (cfg *config) listener() (net.Listener, error) {
	dsn := strings.Split(cfg.Listen, "://")
	if len(dsn) != 2 {
		return nil, errors.New("invalid socket DSN (tcp://:6001, unix://rpc.sock)")
	}

	if dsn[0] == "unix" {
		syscall.Unlink(dsn[1])
	}

	return net.Listen(dsn[0], dsn[1])
}

// dialer creates rpc socket dialer.
func (cfg *config) dialer() (net.Conn, error) {
	dsn := strings.Split(cfg.Listen, "://")
	if len(dsn) != 2 {
		return nil, errors.New("invalid socket DSN (tcp://:6001, unix://rpc.sock)")
	}

	return net.Dial(dsn[0], dsn[1])
}
