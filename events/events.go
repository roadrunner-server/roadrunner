package events

type EventType uint32

const (
	// EventWorkerConstruct thrown when new worker is spawned.
	EventWorkerConstruct EventType = iota
	// EventWorkerDestruct thrown after worker destruction.
	EventWorkerDestruct
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
	// EventWorkerError triggered after WorkerProcess. Except payload to be error.
	EventWorkerError
	// EventWorkerStderr is the worker standard error output
	EventWorkerStderr
	// EventWorkerWaitExit is the worker exit event
	EventWorkerWaitExit
	// EventWorkerStopped triggered when worker gracefully stopped
	EventWorkerStopped
)

func (et EventType) String() string {
	switch et {
	case EventWorkerProcessExit:
		return "EventWorkerProcessExit"
	case EventWorkerConstruct:
		return "EventWorkerConstruct"
	case EventWorkerDestruct:
		return "EventWorkerDestruct"
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
	case EventWorkerError:
		return "EventWorkerError"
	case EventWorkerStderr:
		return "EventWorkerStderr"
	case EventWorkerWaitExit:
		return "EventWorkerWaitExit"
	case EventWorkerStopped:
		return "EventWorkerStopped"
	default:
		return "UnknownEventType"
	}
}
