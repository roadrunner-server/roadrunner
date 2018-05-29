package roadrunner

const (
	// EventWorkerCreate thrown when new worker is spawned.
	EventWorkerCreate = iota

	// EventWorkerDestruct thrown before worker destruction.
	EventWorkerDestruct

	// EventWorkerError thrown any worker related even happen (passed with WorkerError)
	EventWorkerError

	// EventPoolError caused on pool wide errors
	EventPoolError
)

// Pool managed set of inner worker processes.
type Pool interface {
	// Observe all caused events to attached watcher.
	Observe(o func(event int, ctx interface{}))

	// Exec one task with given payload and context, returns result or error.
	Exec(rqs *Payload) (rsp *Payload, err error)

	// Workers returns worker list associated with the pool.
	Workers() (workers []*Worker)

	// Destroy all underlying workers (but let them to complete the task).
	Destroy()
}
