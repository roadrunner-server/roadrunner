// +build linux darwin freebsd

package util

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/valyala/tcplisten"
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

	if dsn[0] == "unix" && fileExists(dsn[1]) {
		err := syscall.Unlink(dsn[1])
		if err != nil {
			return nil, fmt.Errorf("error during the unlink syscall: error %v", err)
		}
	}

	cfg := tcplisten.Config{
		ReusePort:   true,
		DeferAccept: true,
		FastOpen:    true,
		Backlog:     0,
	}

	// tcp4 is currently supported
	if dsn[0] == "tcp" {
		return cfg.NewListener("tcp4", dsn[1])
	}

	return net.Listen(dsn[0], dsn[1])
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
