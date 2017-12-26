package roadrunner

import "time"

// Config defines basic behaviour of worker creation and handling process.
type Config struct {
	// MaxWorkers defines how many sub-processes can be run at once. This value might be doubled by Balancer while hot-swap.
	MaxWorkers uint64

	// MaxExecutions defines how many executions is allowed for the worker until it's destruction. Set 1 to create new process
	// for each new task, 0 to let worker handle as many tasks as it can.
	MaxExecutions uint64

	// AllocateTimeout defines for how long pool will be waiting for a worker to be freed to handle the task.
	AllocateTimeout time.Duration

	// DestroyOnError when set to true workers will be destructed after any JobError.
	DestroyOnError bool
}
