package events

const (
	// EventWorkerConstruct thrown when new worker is spawned.
	EventWorkerConstruct P = iota + 10000

	// EventWorkerDestruct thrown after worker destruction.
	EventWorkerDestruct

	// EventPoolError caused on pool wide errors.
	EventPoolError

	// EventSupervisorError triggered when supervisor can not complete work.
	EventSupervisorError

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
)

type P int64

func (ev P) String() string {
	switch ev {
	case EventWorkerConstruct:
		return "EventWorkerConstruct"
	case EventWorkerDestruct:
		return "EventWorkerDestruct"
	case EventPoolError:
		return "EventPoolError"
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
	}
	return "Unknown event type"
}

// PoolEvent triggered by pool on different events. Pool as also trigger WorkerEvent in case of log.
type PoolEvent struct {
	// Event type, see below.
	Event P

	// Payload depends on event type, typically it's worker or error.
	Payload interface{}
}
