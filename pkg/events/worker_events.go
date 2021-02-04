package events

const (
	// EventWorkerError triggered after WorkerProcess. Except payload to be error.
	EventWorkerError W = iota + 11000
	// EventWorkerLog triggered on every write to WorkerProcess StdErr pipe (batched). Except payload to be []byte string.
	EventWorkerLog
	// EventWorkerStderr is the worker standard error output
	EventWorkerStderr
)

type W int64

func (ev W) String() string {
	switch ev {
	case EventWorkerError:
		return "EventWorkerError"
	case EventWorkerLog:
		return "EventWorkerLog"
	case EventWorkerStderr:
		return "EventWorkerStderr"
	}
	return "Unknown event type"
}

// WorkerEvent wraps worker events.
type WorkerEvent struct {
	// Event id, see below.
	Event W

	// Worker triggered the event.
	Worker interface{}

	// Event specific payload.
	Payload interface{}
}
