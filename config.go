package roadrunner

import (
	"fmt"
	"time"
)

// Config defines basic behaviour of worker creation and handling process.
type Config struct {
	// NumWorkers defines how many sub-processes can be run at once. This value
	// might be doubled by Swapper while hot-swap.
	NumWorkers uint64

	// MaxExecutions defines how many executions is allowed for the worker until
	// it's destruction. set 1 to create new process for each new task, 0 to let
	// worker handle as many tasks as it can.
	MaxExecutions uint64

	// AllocateTimeout defines for how long pool will be waiting for a worker to
	// be freed to handle the task.
	AllocateTimeout time.Duration

	// DestroyTimeout defines for how long pool should be waiting for worker to
	// properly stop, if timeout reached worker will be killed.
	DestroyTimeout time.Duration
}

// Configure returns error if cfg not valid
func (cfg *Config) Valid() error {
	if cfg.NumWorkers == 0 {
		return fmt.Errorf("cfg.NumWorkers must be set")
	}

	if cfg.AllocateTimeout == 0 {
		return fmt.Errorf("cfg.AllocateTimeout must be set")
	}

	if cfg.DestroyTimeout == 0 {
		return fmt.Errorf("cfg.DestroyTimeout must be set")
	}

	return nil
}
