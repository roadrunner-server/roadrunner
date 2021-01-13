package roadrunner

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spiral/roadrunner/osutil"
)

// CommandProducer can produce commands.
type CommandProducer func(cfg *ServerConfig) func() *exec.Cmd

// ServerConfig config combines factory, pool and cmd configurations.
type ServerConfig struct {
	// Command includes command strings with all the parameters, example: "php worker.php pipes".
	Command string

	// User under which process will be started
	User string

	// CommandProducer overwrites
	CommandProducer CommandProducer

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
	mu  sync.Mutex
	env map[string]string
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
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if cfg.env == nil {
		cfg.env = make(map[string]string)
	}

	cfg.env[k] = v
}

// GetEnv must return list of env variables.
func (cfg *ServerConfig) GetEnv() (env []string) {
	env = append(os.Environ(), fmt.Sprintf("RR_RELAY=%s", cfg.Relay))
	for k, v := range cfg.env {
		env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
	}

	return
}

//=================================== PRIVATE METHODS ======================================================

func (cfg *ServerConfig) makeCommand() func() *exec.Cmd {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if cfg.CommandProducer != nil {
		return cfg.CommandProducer(cfg)
	}

	var cmdArgs []string
	cmdArgs = append(cmdArgs, strings.Split(cfg.Command, " ")...)

	return func() *exec.Cmd {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		osutil.IsolateProcess(cmd)

		// if user is not empty, and OS is linux or macos
		// execute php worker from that particular user
		if cfg.User != "" {
			err := osutil.ExecuteFromUser(cmd, cfg.User)
			if err != nil {
				return nil
			}
		}

		cmd.Env = cfg.GetEnv()

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

	if dsn[0] == "unix" && fileExists(dsn[1]) {
		err := syscall.Unlink(dsn[1])
		if err != nil {
			return nil, err
		}
	}

	ln, err := net.Listen(dsn[0], dsn[1])
	if err != nil {
		return nil, err
	}

	return NewSocketFactory(ln, cfg.RelayTimeout), nil
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
