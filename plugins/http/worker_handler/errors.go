// +build !windows

package handler

import (
	"errors"
	"net"
	"os"
	"syscall"
)

// Broken pipe
var errEPIPE = errors.New("EPIPE(32) -> connection reset by peer")

// handleWriteError just check if error was caused by aborted connection on linux
func handleWriteError(err error) error {
	if netErr, ok2 := err.(*net.OpError); ok2 {
		if syscallErr, ok3 := netErr.Err.(*os.SyscallError); ok3 {
			if errors.Is(syscallErr.Err, syscall.EPIPE) {
				return errEPIPE
			}
		}
	}
	return err
}
