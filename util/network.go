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

	switch len(dsn) {
	case 1:
		// socket type prefix not specified
		// assume TCP
		return createTCPListener(dsn[0])
	case 2:
		// socket type prefix specified
		// try parse it
		switch dsn[0] {
		case "unix":
			if fileExists(dsn[1]) {
				err := syscall.Unlink(dsn[1])

				if err != nil {
					return nil, fmt.Errorf("error during the unlink syscall: error %v", err)
				}
			}

			return net.Listen(dsn[0], dsn[1])
		case "tcp":
			return createTCPListener(dsn[1])
		default:
			// fail out if invalid socket type is given
			return nil, errors.New("invalid Protocol ([tcp://]:6001, unix://file.sock)")
		}
	default:
		// invalid syntax
		return nil, errors.New("invalid DSN ([tcp://]:6001, unix://file.sock)")
	}
}

func createTCPListener(addr string) (net.Listener, error) {
	cfg := tcplisten.Config{
		ReusePort:   true,
		DeferAccept: true,
		FastOpen:    true,
		Backlog:     0,
	}

	listener, err := cfg.NewListener("tcp4", addr)

	if err != nil {
		return nil, err
	}

	return listener, nil;
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
