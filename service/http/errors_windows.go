// +build windows

package http

import (
	"errors"
	"net"
	"os"
	"syscall"
)

var errEPIPE = errors.New("WSAECONNABORTED (10053) ->  an established connection was aborted")

// handleWriteError just check if error was caused by aborted connection on windows
func handleWriteError(err error) error {
	if netErr, ok2 := err.(*net.OpError); ok2 {
		if syscallErr, ok3 := netErr.Err.(*os.SyscallError); ok3 {
			if syscallErr.Err == syscall.WSAECONNABORTED {
				return errEPIPE
			}
		}
	}
	return err
}