package roadrunner

const (
	// EventCreated thrown when new worker is spawned.
	EventCreated = iota

	// EventDestruct thrown before worker destruction.
	EventDestruct

	// EventError thrown any worker related even happen (error passed as context)
	EventError
)

// Pool managed set of inner worker processes.
type Pool interface {
	// Report all caused events to attached watcher.
	Report(o func(event int, w *Worker, ctx interface{}))

	// Exec one task with given payload and context, returns result or error.
	Exec(rqs *Payload) (rsp *Payload, err error)

	// Workers returns worker list associated with the pool.
	Workers() (workers []*Worker)

	// Destroy all underlying workers (but let them to complete the task).
	Destroy()
}
