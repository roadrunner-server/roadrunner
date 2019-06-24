package util

import (
	"errors"
	"net"
	"strings"
	"syscall"
)

// CreateListener crates socket listener based on DSN definition.
func CreateListener(address string) (net.Listener, error) {
	dsn := strings.Split(address, "://")
	if len(dsn) != 2 {
		return nil, errors.New("Invalid DSN (tcp://:6001, unix://file.sock)")
	}

	if dsn[0] != "unix" && dsn[0] != "tcp" {
		return nil, errors.New("Invalid Protocol (tcp://:6001, unix://file.sock)")
	}

	if dsn[0] == "unix" {
		syscall.Unlink(dsn[1])
	}

	return net.Listen(dsn[0], dsn[1])
}
