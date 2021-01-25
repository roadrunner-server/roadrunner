package worker_watcher //nolint:golint,stylecheck

import (
	"context"

	"github.com/spiral/roadrunner/v2/pkg/worker"
)

type Watcher interface {
	// AddToWatch used to add stack to wait its state
	AddToWatch(workers []worker.SyncWorker) error

	// GetFreeWorker provide first free worker
	GetFreeWorker(ctx context.Context) (worker.SyncWorker, error)

	// PutWorker enqueues worker back
	PushWorker(w worker.SyncWorker)

	// AllocateNew used to allocate new worker and put in into the WorkerWatcher
	AllocateNew() error

	// Destroy destroys the underlying stack
	Destroy(ctx context.Context)

	// WorkersList return all stack w/o removing it from internal storage
	WorkersList() []worker.SyncWorker

	// RemoveWorker remove worker from the stack
	RemoveWorker(wb worker.SyncWorker) error
}
