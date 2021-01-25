// +build windows

package http

import (
	"errors"
	"net"
	"os"
	"syscall"
)

//Software caused connection abort.
//An established connection was aborted by the software in your host computer,
//possibly due to a data transmission time-out or protocol error.
var errEPIPE = errors.New("WSAECONNABORTED (10053) ->  an established connection was aborted by peer")

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
