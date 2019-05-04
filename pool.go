package roadrunner

const (
	// EventWorkerConstruct thrown when new worker is spawned.
	EventWorkerConstruct = iota + 100

	// EventWorkerDestruct thrown after worker destruction.
	EventWorkerDestruct

	// EventWorkerKill thrown after worker is being forcefully killed.
	EventWorkerKill

	// EventWorkerError thrown any worker related even happen (passed with WorkerError)
	EventWorkerError

	// EventWorkerDead thrown when worker stops worker for any reason.
	EventWorkerDead

	// EventPoolError caused on pool wide errors
	EventPoolError
)

// Pool managed set of inner worker processes.
type Pool interface {
	// Listen all caused events to attached controller.
	Listen(l func(event int, ctx interface{}))

	// Exec one task with given payload and context, returns result or error.
	Exec(rqs *Payload) (rsp *Payload, err error)

	// Workers returns worker list associated with the pool.
	Workers() (workers []*Worker)

	// Remove forces pool to remove specific worker. Return true is this is first remove request on given worker.
	Remove(w *Worker, err error) bool

	// Destroy all underlying workers (but let them to complete the task).
	Destroy()
}
