package roadrunner

import (
	"context"
	"runtime"
	"time"
)

const (
	// EventWorkerConstruct thrown when new worker is spawned.
	EventWorkerConstruct = iota + 100

	// EventWorkerDestruct thrown after worker destruction.
	EventWorkerDestruct

	// EventWorkerKill thrown after worker is being forcefully killed.
	EventWorkerKill

	// EventWorkerError thrown any worker related even happen (passed with WorkerError)
	EventWorkerEvent

	// EventWorkerDead thrown when worker stops worker for any reason.
	EventWorkerDead

	// EventPoolError caused on pool wide errors
	EventPoolError
)

const (
	// EventMaxMemory caused when worker consumes more memory than allowed.
	EventMaxMemory = iota + 8000

	// EventTTL thrown when worker is removed due TTL being reached. Context is rr.WorkerError
	EventTTL

	// EventIdleTTL triggered when worker spends too much time at rest.
	EventIdleTTL

	// EventExecTTL triggered when worker spends too much time doing the task (max_execution_time).
	EventExecTTL
)

// Pool managed set of inner worker processes.
type Pool interface {
	// ATTENTION, YOU SHOULD CONSUME EVENTS, OTHERWISE POOL WILL BLOCK
	Events() chan PoolEvent

	// Exec one task with given payload and context, returns result or error.
	ExecWithContext(ctx context.Context, rqs Payload) (Payload, error)

	// Exec
	Exec(rqs Payload) (Payload, error)

	// Workers returns worker list associated with the pool.
	Workers() (workers []WorkerBase)

	RemoveWorker(ctx context.Context, worker WorkerBase) error

	Config() Config

	// Destroy all underlying stack (but let them to complete the task).
	Destroy(ctx context.Context)
}

// todo: merge with pool options

// Config defines basic behaviour of worker creation and handling process.
//
type Config struct {
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

	// TTL defines maximum time worker is allowed to live.
	TTL int64

	// IdleTTL defines maximum duration worker can spend in idle mode. Disabled when 0.
	IdleTTL int64

	// ExecTTL defines maximum lifetime per job.
	ExecTTL time.Duration

	// MaxPoolMemory defines maximum amount of memory allowed for worker. In megabytes.
	MaxPoolMemory uint64

	// MaxWorkerMemory limits memory per worker.
	MaxWorkerMemory uint64
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
}
