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
	NumWorkers int64

	// MaxJobs defines how many executions is allowed for the worker until
	// it's destruction. set 1 to create new process for each new task, 0 to let
	// worker handle as many tasks as it can.
	MaxJobs int64

	// AllocateTimeout defines for how long pool will be waiting for a worker to
	// be freed to handle the task. Defaults to 60s.
	AllocateTimeout time.Duration

	// DestroyTimeout defines for how long pool should be waiting for worker to
	// properly destroy, if timeout reached worker will be killed. Defaults to 60s.
	DestroyTimeout time.Duration

	// Supervision config to limit worker and pool memory usage.
	Supervisor *SupervisorConfig
}

// InitDefaults enables default config values.
func (cfg *Config) InitDefaults() {
	if cfg.NumWorkers == 0 {
		cfg.NumWorkers = int64(runtime.NumCPU())
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
	WatchTick uint64

	// TTL defines maximum time worker is allowed to live.
	TTL uint64

	// IdleTTL defines maximum duration worker can spend in idle mode. Disabled when 0.
	IdleTTL uint64

	// ExecTTL defines maximum lifetime per job.
	ExecTTL uint64

	// MaxWorkerMemory limits memory per worker.
	MaxWorkerMemory uint64
}

// InitDefaults enables default config values.
func (cfg *SupervisorConfig) InitDefaults() {
	if cfg.WatchTick == 0 {
		cfg.WatchTick = 1
	}
}
