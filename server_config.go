package roadrunner

import (
	"errors"
	"net"
	"strings"
	"time"
	"os/exec"
	"syscall"
	"os/user"
	"strconv"
)

const (
	FactoryPipes  = iota
	FactorySocket
)

// Server config combines factory, pool and cmd configurations.
type ServerConfig struct {
	// Command includes command strings with all the parameters, example: "php worker.php pipes". This config section
	//	// must not change on re-configuration.
	Command string

	// User specifies what user to run command under, for Unix systems only. Support both UID and name options. Keep
	// empty to use current user.This config section must not change on re-configuration.
	User string

	// Group specifies what group to run command under, for Unix systems only. Support GID or name options. Keep empty
	// to use current user.This config section must not change on re-configuration.
	Group string

	// Relay defines connection method and factory to be used to connect to workers:
	// "pipes", "tcp://:6001", "unix://rr.sock"
	// This config section must not change on re-configuration.
	Relay string

	// FactoryTimeout defines for how long socket factory will be waiting for worker connection. This config section
	// must not change on re-configuration.
	FactoryTimeout time.Duration

	// Pool defines worker pool configuration, number of workers, timeouts and etc. This config section might change
	// while server is running.
	Pool *Config
}

func (cfg *ServerConfig) makeCommand() (func() *exec.Cmd, error) {
	var (
		err error
		u   *user.User
		g   *user.Group
		crd *syscall.Credential
		cmd = strings.Split(cfg.Command, " ")
	)

	if cfg.User != "" {
		if u, err = resolveUser(cfg.User); err != nil {
			return nil, err
		}
	}

	if cfg.Group != "" {
		if g, err = resolveGroup(cfg.Group); err != nil {
			return nil, err
		}
	}

	if u != nil || g != nil {
		crd = &syscall.Credential{}

		if u != nil {
			uid, err := strconv.ParseUint(u.Uid, 10, 32)
			if err != nil {
				return nil, err
			}

			crd.Uid = uint32(uid)
		}

		if g != nil {
			gid, err := strconv.ParseUint(g.Gid, 10, 32)
			if err != nil {
				return nil, err
			}

			crd.Gid = uint32(gid)
		}
	}

	return func() *exec.Cmd {
		cmd := exec.Command(cmd[0], cmd[1:]...)
		if crd != nil {
			cmd.SysProcAttr = &syscall.SysProcAttr{Credential: crd}
		}

		return cmd
	}, nil
}

// makeFactory creates and connects new factory instance based on given parameters.
func (cfg *ServerConfig) makeFactory() (Factory, error) {
	if cfg.Relay == "pipes" || cfg.Relay == "pipe" {
		return NewPipeFactory(), nil
	}

	dsn := strings.Split(cfg.Relay, "://")
	if len(dsn) != 2 {
		return nil, errors.New("invalid relay DSN (pipes, tcp://:6001, unix://rr.sock)")
	}

	ln, err := net.Listen(dsn[0], dsn[1])
	if err != nil {
		return nil, nil
	}

	return NewSocketFactory(ln, time.Second*cfg.FactoryTimeout), nil
}
