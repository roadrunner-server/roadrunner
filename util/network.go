package util

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"
)

// CreateListener crates socket listener based on DSN definition.
func CreateListener(address string) (net.Listener, error) {
	dsn := strings.Split(address, "://")
	if len(dsn) != 2 {
		return nil, errors.New("invalid DSN (tcp://:6001, unix://file.sock)")
	}

	if dsn[0] != "unix" && dsn[0] != "tcp" {
		return nil, errors.New("invalid Protocol (tcp://:6001, unix://file.sock)")
	}

	if dsn[0] == "unix" {
		err := syscall.Unlink(dsn[1])
		if err != nil {
			return nil, fmt.Errorf("error during the unlink syscall: error %v", err)
		}
	}

	return net.Listen(dsn[0], dsn[1])
}
