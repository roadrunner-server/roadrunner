package events

// EventWorkerKill thrown after WorkerProcess is being forcefully killed.
const (
	// EventWorkerError triggered after WorkerProcess. Except payload to be error.
	EventWorkerError E = iota + 200

	// EventWorkerLog triggered on every write to WorkerProcess StdErr pipe (batched). Except payload to be []byte string.
	EventWorkerLog
)

type E int64

func (ev E) String() string {
	switch ev {
	case EventWorkerError:
		return "EventWorkerError"
	case EventWorkerLog:
		return "EventWorkerLog"
	}
	return "Unknown event type"
}

// WorkerEvent wraps worker events.
type WorkerEvent struct {
	// Event id, see below.
	Event E

	// Worker triggered the event.
	Worker interface{}

	// Event specific payload.
	Payload interface{}
}
