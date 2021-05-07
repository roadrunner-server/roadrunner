// +build !windows

package osutil

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

// IsolateProcess change gpid for the process to avoid bypassing signals to php processes.
func IsolateProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}
}

// ExecuteFromUser may work only if run RR under root user
func ExecuteFromUser(cmd *exec.Cmd, u string) error {
	usr, err := user.Lookup(u)
	if err != nil {
		return err
	}

	usrI32, err := strconv.Atoi(usr.Uid)
	if err != nil {
		return err
	}

	grI32, err := strconv.Atoi(usr.Gid)
	if err != nil {
		return err
	}

	// For more information:
	// https://www.man7.org/linux/man-pages/man7/user_namespaces.7.html
	// https://www.man7.org/linux/man-pages/man7/namespaces.7.html
	if _, err := os.Stat("/proc/self/ns/user"); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("kernel doesn't support user namespaces")
		}
		if os.IsPermission(err) {
			return fmt.Errorf("unable to test user namespaces due to permissions")
		}

		return fmt.Errorf("failed to stat /proc/self/ns/user: %v", err)
	}

	cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid: uint32(usrI32),
		Gid: uint32(grI32),
	}

	return nil
}
