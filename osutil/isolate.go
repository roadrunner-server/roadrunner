// +build !windows

package osutil

import (
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

// IsolateProcess change gpid for the process to avoid bypassing signals to php processes.
func IsolateProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}
}

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

	cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid: uint32(usrI32),
		Gid: uint32(grI32),
	}

	return nil
}
