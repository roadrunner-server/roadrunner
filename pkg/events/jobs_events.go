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

	// EventPipeConsume when pipeline pipelines has been requested.
	EventPipeConsume

	// EventPipeActive when pipeline has started.
	EventPipeActive

	// EventPipeStop when pipeline has begun stopping.
	EventPipeStop

	// EventPipeStopped when pipeline has been stopped.
	EventPipeStopped

	// EventPipeError when pipeline specific error happen.
	EventPipeError

	// EventBrokerReady thrown when broken is ready to accept/serve tasks.
	EventBrokerReady
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
	case EventPipeConsume:
		return "EventPipeConsume"
	case EventPipeActive:
		return "EventPipeActive"
	case EventPipeStop:
		return "EventPipeStop"
	case EventPipeStopped:
		return "EventPipeStopped"
	case EventPipeError:
		return "EventPipeError"
	case EventBrokerReady:
		return "EventBrokerReady"
	}
	return UnknownEventType
}

// JobEvent represent job event.
type JobEvent struct {
	Event J
	// String is job id.
	ID string

	// Job is failed job.
	Job interface{} // this is *jobs.Job, but interface used to avoid package import

	// event timings
	Start   time.Time
	Elapsed time.Duration
}
