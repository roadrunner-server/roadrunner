package events

import (
	"time"

	"google.golang.org/grpc"
)

const (
	// EventUnaryCallOk represents success unary call response
	EventUnaryCallOk G = iota + 13000

	// EventUnaryCallErr raised when unary call ended with error
	EventUnaryCallErr
)

type G int64

func (ev G) String() string {
	switch ev {
	case EventUnaryCallOk:
		return "EventUnaryCallOk"
	case EventUnaryCallErr:
		return "EventUnaryCallErr"
	}
	return UnknownEventType
}

// JobEvent represent job event.
type GRPCEvent struct {
	Event G
	// Info contains unary call info.
	Info *grpc.UnaryServerInfo
	// Error associated with event.
	Error error
	// event timings
	Start   time.Time
	Elapsed time.Duration
}
