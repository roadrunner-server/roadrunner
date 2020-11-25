// +build linux darwin freebsd

package util

import (
	"errors"
	"fmt"
	"github.com/valyala/tcplisten"
	"net"
	"os"
	"strings"
	"syscall"
)

// CreateListener crates socket listener based on DSN definition.
func CreateListener(address string) (net.Listener, error) {
	dsn := strings.Split(address, "://")

	var protocol string
	var param    string

	if len(dsn) == 1 {
		protocol = "tcp"
		param = dsn[0]
	} else if len(dsn) == 2 {
		if dsn[0] != "unix" && dsn[0] != "tcp" {
			return nil, errors.New("invalid Protocol ([tcp://]:6001, unix://file.sock)")
		} else {
			protocol = dsn[0]
			param = dsn[1]
		}
	} else {
		return nil, errors.New("invalid DSN ([tcp://]:6001, unix://file.sock)")
	}

	if protocol == "unix" && fileExists(param) {
		err := syscall.Unlink(param)
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
	if protocol == "tcp" {
		return cfg.NewListener("tcp4", param)
	}

	return net.Listen(protocol, param)
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
