package events

import (
	"time"
)

const (
	// EventPushOK thrown when new job has been added. JobEvent is passed as context.
	EventPushOK = iota + 12000

	// EventPushError caused when job can not be registered.
	EventPushError

	// EventJobStart thrown when new job received.
	EventJobStart

	// EventJobOK thrown when job execution is successfully completed. JobEvent is passed as context.
	EventJobOK

	// EventJobError thrown on all job related errors. See JobError as context.
	EventJobError

	// EventPipeActive when pipeline has started.
	EventPipeActive

	// EventPipeStopped when pipeline has been stopped.
	EventPipeStopped

	// EventPipePaused when pipeline has been paused.
	EventPipePaused

	// EventPipeError when pipeline specific error happen.
	EventPipeError

	// EventDriverReady thrown when broken is ready to accept/serve tasks.
	EventDriverReady
)

type J int64

func (ev J) String() string {
	switch ev {
	case EventPushOK:
		return "EventPushOK"
	case EventPushError:
		return "EventPushError"
	case EventJobStart:
		return "EventJobStart"
	case EventJobOK:
		return "EventJobOK"
	case EventJobError:
		return "EventJobError"
	case EventPipeActive:
		return "EventPipeActive"
	case EventPipeStopped:
		return "EventPipeStopped"
	case EventPipeError:
		return "EventPipeError"
	case EventDriverReady:
		return "EventDriverReady"
	}
	return UnknownEventType
}

// JobEvent represent job event.
type JobEvent struct {
	Event J
	// String is job id.
	ID string

	// Pipeline name
	Pipeline string

	// Associated driver name (amqp, ephemeral, etc)
	Driver string

	// Error for the jobs/pipes errors
	Error error

	// event timings
	Start   time.Time
	Elapsed time.Duration
}
