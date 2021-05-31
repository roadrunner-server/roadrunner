package pool

import (
	"context"

	"github.com/spiral/roadrunner/v2/pkg/payload"
	"github.com/spiral/roadrunner/v2/pkg/worker"
)

// Pool managed set of inner worker processes.
type Pool interface {
	// GetConfig returns pool configuration.
	GetConfig() interface{}

	// Exec executes task with payload
	Exec(rqs payload.Payload) (payload.Payload, error)

	// Workers returns worker list associated with the pool.
	Workers() (workers []worker.BaseProcess)

	// RemoveWorker removes worker from the pool.
	RemoveWorker(worker worker.BaseProcess) error

	// Destroy all underlying stack (but let them to complete the task).
	Destroy(ctx context.Context)

	// ExecWithContext executes task with context which is used with timeout
	execWithTTL(ctx context.Context, rqs payload.Payload) (payload.Payload, error)
}
