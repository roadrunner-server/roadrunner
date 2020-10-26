package roadrunner

import (
	"context"
	"runtime"
	"time"

	"github.com/spiral/roadrunner/v2/util"
)

// PoolEvent triggered by pool on different events. Pool as also trigger WorkerEvent in case of log.
type PoolEvent struct {
	// Event type, see below.
	Event int64

	// Payload depends on event type, typically it's worker or error.
	Payload interface{}
}

const (
	// EventWorkerConstruct thrown when new worker is spawned.
	EventWorkerConstruct = iota + 7800

	// EventWorkerDestruct thrown after worker destruction.
	EventWorkerDestruct

	// EventPoolError caused on pool wide errors.
	EventPoolError

	// EventSupervisorError triggered when supervisor can not complete work.
	EventSupervisorError

	// todo: EventMaxMemory caused when worker consumes more memory than allowed.
	EventMaxMemory

	// todo: EventTTL thrown when worker is removed due TTL being reached. Context is rr.WorkerError
	EventTTL

	// todo: EventIdleTTL triggered when worker spends too much time at rest.
	EventIdleTTL

	// todo: EventExecTTL triggered when worker spends too much time doing the task (max_execution_time).
	EventExecTTL
)

// Pool managed set of inner worker processes.
type Pool interface {
	// AddListener connects event listener to the pool.
	AddListener(listener util.EventListener)

	// GetConfig returns pool configuration.
	GetConfig() Config

	// Exec
	Exec(rqs Payload) (Payload, error)

	// Workers returns worker list associated with the pool.
	Workers() (workers []WorkerBase)

	// Remove worker from the pool.
	RemoveWorker(ctx context.Context, worker WorkerBase) error

	// Destroy all underlying stack (but let them to complete the task).
	Destroy(ctx context.Context)
}

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
	Supervisor SupervisorConfig
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

	cfg.Supervisor.InitDefaults()
}

type SupervisorConfig struct {
	// WatchTick defines how often to check the state of worker.
	WatchTick time.Duration

	// TTL defines maximum time worker is allowed to live.
	TTL int64

	// IdleTTL defines maximum duration worker can spend in idle mode. Disabled when 0.
	IdleTTL int64

	// ExecTTL defines maximum lifetime per job.
	ExecTTL time.Duration

	// MaxWorkerMemory limits memory per worker.
	MaxWorkerMemory uint64
}

// InitDefaults enables default config values.
func (cfg *SupervisorConfig) InitDefaults() {
	if cfg.WatchTick == 0 {
		cfg.WatchTick = time.Second
	}
}
