package internal

import (
	"fmt"
	"sync/atomic"

	"github.com/spiral/roadrunner/v2/pkg/states"
)

// State represents WorkerProcess status and updated time.
type State interface {
	fmt.Stringer
	// Value returns WorkerState value
	Value() int64
	// Set sets the WorkerState
	Set(value int64)
	// NumJobs shows how many times WorkerProcess was invoked
	NumExecs() uint64
	// IsActive returns true if WorkerProcess not Inactive or Stopped
	IsActive() bool
	// RegisterExec using to registering php executions
	RegisterExec()
	// SetLastUsed sets worker last used time
	SetLastUsed(lu uint64)
	// LastUsed return worker last used time
	LastUsed() uint64
}

type WorkerState struct {
	value    int64
	numExecs uint64
	// to be lightweight, use UnixNano
	lastUsed uint64
}

// Thread safe
func NewWorkerState(value int64) *WorkerState {
	return &WorkerState{value: value}
}

// String returns current WorkerState as string.
func (s *WorkerState) String() string {
	switch s.Value() {
	case states.StateInactive:
		return "inactive"
	case states.StateReady:
		return "ready"
	case states.StateWorking:
		return "working"
	case states.StateInvalid:
		return "invalid"
	case states.StateStopped:
		return "stopped"
	case states.StateErrored:
		return "errored"
	}

	return "undefined"
}

// NumExecs returns number of registered WorkerProcess execs.
func (s *WorkerState) NumExecs() uint64 {
	return atomic.LoadUint64(&s.numExecs)
}

// Value WorkerState returns WorkerState value
func (s *WorkerState) Value() int64 {
	return atomic.LoadInt64(&s.value)
}

// IsActive returns true if WorkerProcess not Inactive or Stopped
func (s *WorkerState) IsActive() bool {
	val := s.Value()
	return val == states.StateWorking || val == states.StateReady
}

// change WorkerState value (status)
func (s *WorkerState) Set(value int64) {
	atomic.StoreInt64(&s.value, value)
}

// register new execution atomically
func (s *WorkerState) RegisterExec() {
	atomic.AddUint64(&s.numExecs, 1)
}

// Update last used time
func (s *WorkerState) SetLastUsed(lu uint64) {
	atomic.StoreUint64(&s.lastUsed, lu)
}

func (s *WorkerState) LastUsed() uint64 {
	return atomic.LoadUint64(&s.lastUsed)
}
