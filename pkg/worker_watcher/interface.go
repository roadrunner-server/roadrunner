package worker_watcher //nolint:stylecheck

import (
	"context"

	"github.com/spiral/roadrunner/v2/pkg/worker"
)

// Watcher is an interface for the Sync workers lifecycle
type Watcher interface {
	// Watch used to add workers to the container
	Watch(workers []worker.BaseProcess) error

	// Get provide first free worker
	Get(ctx context.Context) (worker.BaseProcess, error)

	// Push enqueues worker back
	Push(w worker.BaseProcess)

	// Allocate - allocates new worker and put it into the WorkerWatcher
	Allocate() error

	// Destroy destroys the underlying container
	Destroy(ctx context.Context)

	// List return all container w/o removing it from internal storage
	List() []worker.BaseProcess

	// Remove will remove worker from the container
	Remove(wb worker.BaseProcess)
}
