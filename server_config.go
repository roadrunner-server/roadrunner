package roadrunner

import (
	"errors"
	"fmt"
	"github.com/spiral/roadrunner/osutil"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// ServerConfig config combines factory, pool and cmd configurations.
type ServerConfig struct {
	// Command includes command strings with all the parameters, example: "php worker.php pipes".
	Command string

	// Relay defines connection method and factory to be used to connect to workers:
	// "pipes", "tcp://:6001", "unix://rr.sock"
	// This config section must not change on re-configuration.
	Relay string

	// RelayTimeout defines for how long socket factory will be waiting for worker connection. This config section
	// must not change on re-configuration.
	RelayTimeout time.Duration

	// Pool defines worker pool configuration, number of workers, timeouts and etc. This config section might change
	// while server is running.
	Pool *Config

	// values defines set of values to be passed to the command context.
	env []string
}

// InitDefaults sets missing values to their default values.
func (cfg *ServerConfig) InitDefaults() error {
	cfg.Relay = "pipes"
	cfg.RelayTimeout = time.Minute

	if cfg.Pool == nil {
		cfg.Pool = &Config{}
	}

	return cfg.Pool.InitDefaults()
}

// UpscaleDurations converts duration values from nanoseconds to seconds.
func (cfg *ServerConfig) UpscaleDurations() {
	if cfg.RelayTimeout < time.Microsecond {
		cfg.RelayTimeout = time.Second * time.Duration(cfg.RelayTimeout.Nanoseconds())
	}

	if cfg.Pool.AllocateTimeout < time.Microsecond {
		cfg.Pool.AllocateTimeout = time.Second * time.Duration(cfg.Pool.AllocateTimeout.Nanoseconds())
	}

	if cfg.Pool.DestroyTimeout < time.Microsecond {
		cfg.Pool.DestroyTimeout = time.Second * time.Duration(cfg.Pool.DestroyTimeout.Nanoseconds())
	}
}

// Differs returns true if configuration has changed but ignores pool or cmd changes.
func (cfg *ServerConfig) Differs(new *ServerConfig) bool {
	return cfg.Relay != new.Relay || cfg.RelayTimeout != new.RelayTimeout
}

// SetEnv sets new environment variable. Value is automatically uppercase-d.
func (cfg *ServerConfig) SetEnv(k, v string) {
	cfg.env = append(cfg.env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
}

// makeCommands returns new command provider based on configured options.
func (cfg *ServerConfig) makeCommand() func() *exec.Cmd {
	var cmd = strings.Split(cfg.Command, " ")
	return func() *exec.Cmd {
		cmd := exec.Command(cmd[0], cmd[1:]...)
		osutil.IsolateProcess(cmd)

		cmd.Env = append(os.Environ(), fmt.Sprintf("RR_RELAY=%s", cfg.Relay))
		cmd.Env = append(cmd.Env, cfg.env...)
		return cmd
	}
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

	if dsn[0] == "unix" {
		syscall.Unlink(dsn[1])
	}

	ln, err := net.Listen(dsn[0], dsn[1])
	if err != nil {
		return nil, err
	}

	return NewSocketFactory(ln, cfg.RelayTimeout), nil
}
