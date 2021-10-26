package events

import (
	"fmt"
)

type EventBus interface {
	SubscribeAll(subID string, ch chan<- Event) error
	SubscribeP(subID string, pattern string, ch chan<- Event) error
	Unsubscribe(subID string)
	UnsubscribeP(subID, pattern string)
	Send(ev Event)
}

type Event interface {
	fmt.Stringer
	Plugin() string
	Type() EventType
	Message() string
}

type RREvent struct {
	// event typ
	T EventType
	// plugin
	P string
	// message
	M string
}

// NewRREvent initializes new event
func NewRREvent(t EventType, msg string, plugin string) *RREvent {
	return &RREvent{
		T: t,
		P: plugin,
		M: msg,
	}
}

func (r *RREvent) String() string {
	return "RoadRunner event"
}

func (r *RREvent) Type() EventType {
	return r.T
}

func (r *RREvent) Message() string {
	return r.M
}

func (r *RREvent) Plugin() string {
	return r.P
}

type EventType uint32

const (
	// EventUnaryCallOk represents success unary call response
	EventUnaryCallOk EventType = iota

	// EventUnaryCallErr raised when unary call ended with error
	EventUnaryCallErr

	// EventPushOK thrown when new job has been added. JobEvent is passed as context.
	EventPushOK

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

	// EventWorkerConstruct thrown when new worker is spawned.
	EventWorkerConstruct

	// EventWorkerDestruct thrown after worker destruction.
	EventWorkerDestruct

	// EventSupervisorError triggered when supervisor can not complete work.
	EventSupervisorError

	// EventWorkerProcessExit triggered on process wait exit
	EventWorkerProcessExit

	// EventNoFreeWorkers triggered when there are no free workers in the stack and timeout for worker allocate elapsed
	EventNoFreeWorkers

	// EventMaxMemory caused when worker consumes more memory than allowed.
	EventMaxMemory

	// EventTTL thrown when worker is removed due TTL being reached. TTL defines maximum time worker is allowed to live (seconds)
	EventTTL

	// EventIdleTTL triggered when worker spends too much time at rest.
	EventIdleTTL

	// EventExecTTL triggered when worker spends too much time doing the task (max_execution_time).
	EventExecTTL

	// EventPoolRestart triggered when pool restart is needed
	EventPoolRestart

	// EventWorkerError triggered after WorkerProcess. Except payload to be error.
	EventWorkerError
	// EventWorkerLog triggered on every write to WorkerProcess StdErr pipe (batched). Except payload to be []byte string.
	EventWorkerLog
	// EventWorkerStderr is the worker standard error output
	EventWorkerStderr
	// EventWorkerWaitExit is the worker exit event
	EventWorkerWaitExit
)

func (et EventType) String() string {
	switch et {
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
	case EventPipePaused:
		return "EventPipePaused"

	case EventUnaryCallOk:
		return "EventUnaryCallOk"
	case EventUnaryCallErr:
		return "EventUnaryCallErr"

	case EventWorkerProcessExit:
		return "EventWorkerProcessExit"
	case EventWorkerConstruct:
		return "EventWorkerConstruct"
	case EventWorkerDestruct:
		return "EventWorkerDestruct"
	case EventSupervisorError:
		return "EventSupervisorError"
	case EventNoFreeWorkers:
		return "EventNoFreeWorkers"
	case EventMaxMemory:
		return "EventMaxMemory"
	case EventTTL:
		return "EventTTL"
	case EventIdleTTL:
		return "EventIdleTTL"
	case EventExecTTL:
		return "EventExecTTL"
	case EventPoolRestart:
		return "EventPoolRestart"

	case EventWorkerError:
		return "EventWorkerError"
	case EventWorkerLog:
		return "EventWorkerLog"
	case EventWorkerStderr:
		return "EventWorkerStderr"
	case EventWorkerWaitExit:
		return "EventWorkerWaitExit"

	default:
		return "UnknownEventType"
	}
}
