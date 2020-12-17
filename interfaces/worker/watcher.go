package worker

import "context"

type Watcher interface {
	// AddToWatch used to add stack to wait its state
	AddToWatch(workers []BaseProcess) error

	// GetFreeWorker provide first free worker
	GetFreeWorker(ctx context.Context) (BaseProcess, error)

	// PutWorker enqueues worker back
	PushWorker(w BaseProcess)

	// AllocateNew used to allocate new worker and put in into the WorkerWatcher
	AllocateNew() error

	// Destroy destroys the underlying stack
	Destroy(ctx context.Context)

	// WorkersList return all stack w/o removing it from internal storage
	WorkersList() []BaseProcess

	// RemoveWorker remove worker from the stack
	RemoveWorker(wb BaseProcess) error
}
