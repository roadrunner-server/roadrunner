package pool

import (
	"runtime"
	"time"
)

// Configures the pool behaviour.
type Config struct {
	// Debug flag creates new fresh worker before every request.
	Debug bool

	// NumWorkers defines how many sub-processes can be run at once. This value
	// might be doubled by Swapper while hot-swap. Defaults to number of CPU cores.
	NumWorkers uint64 `mapstructure:"num_workers"`

	// MaxJobs defines how many executions is allowed for the worker until
	// it's destruction. set 1 to create new process for each new task, 0 to let
	// worker handle as many tasks as it can.
	MaxJobs uint64 `mapstructure:"max_jobs"`

	// AllocateTimeout defines for how long pool will be waiting for a worker to
	// be freed to handle the task. Defaults to 60s.
	AllocateTimeout time.Duration `mapstructure:"allocate_timeout"`

	// DestroyTimeout defines for how long pool should be waiting for worker to
	// properly destroy, if timeout reached worker will be killed. Defaults to 60s.
	DestroyTimeout time.Duration `mapstructure:"destroy_timeout"`

	// Supervision config to limit worker and pool memory usage.
	Supervisor *SupervisorConfig `mapstructure:"supervisor"`
}

// InitDefaults enables default config values.
func (cfg *Config) InitDefaults() {
	if cfg.NumWorkers == 0 {
		cfg.NumWorkers = uint64(runtime.NumCPU())
	}

	if cfg.AllocateTimeout == 0 {
		cfg.AllocateTimeout = time.Minute
	}

	if cfg.DestroyTimeout == 0 {
		cfg.DestroyTimeout = time.Minute
	}
	if cfg.Supervisor == nil {
		return
	}
	cfg.Supervisor.InitDefaults()
}

type SupervisorConfig struct {
	// WatchTick defines how often to check the state of worker.
	WatchTick time.Duration `mapstructure:"watch_tick"`

	// TTL defines maximum time worker is allowed to live.
	TTL time.Duration `mapstructure:"ttl"`

	// IdleTTL defines maximum duration worker can spend in idle mode. Disabled when 0.
	IdleTTL time.Duration `mapstructure:"idle_ttl"`

	// ExecTTL defines maximum lifetime per job.
	ExecTTL time.Duration `mapstructure:"exec_ttl"`

	// MaxWorkerMemory limits memory per worker.
	MaxWorkerMemory uint64 `mapstructure:"max_worker_memory"`
}

// InitDefaults enables default config values.
func (cfg *SupervisorConfig) InitDefaults() {
	if cfg.WatchTick == 0 {
		cfg.WatchTick = time.Second
	}
}
