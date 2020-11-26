package roadrunner

import (
	"fmt"
	"sync/atomic"
)

// State represents WorkerProcess status and updated time.
type State interface {
	fmt.Stringer

	// Value returns state value
	Value() int64
	// Set sets the state
	Set(value int64)
	// NumJobs shows how many times WorkerProcess was invoked
	NumExecs() int64
	// IsActive returns true if WorkerProcess not Inactive or Stopped
	IsActive() bool
	// RegisterExec using to registering php executions
	RegisterExec()
	// SetLastUsed sets worker last used time
	SetLastUsed(lu uint64)
	// LastUsed return worker last used time
	LastUsed() uint64
}

const (
	// StateInactive - no associated process
	StateInactive int64 = iota

	// StateReady - ready for job.
	StateReady

	// StateWorking - working on given payload.
	StateWorking

	// StateInvalid - indicates that WorkerProcess is being disabled and will be removed.
	StateInvalid

	// StateStopping - process is being softly stopped.
	StateStopping

	StateKilling

	// State of worker, when no need to allocate new one
	StateDestroyed

	// StateStopped - process has been terminated.
	StateStopped

	// StateErrored - error state (can't be used).
	StateErrored

	StateRemove
)

type state struct {
	value    int64
	numExecs int64
	// to be lightweight, use UnixNano
	lastUsed uint64
}

// Thread safe
func newState(value int64) *state {
	return &state{value: value}
}

// String returns current state as string.
func (s *state) String() string {
	switch s.Value() {
	case StateInactive:
		return "inactive"
	case StateReady:
		return "ready"
	case StateWorking:
		return "working"
	case StateInvalid:
		return "invalid"
	case StateStopped:
		return "stopped"
	case StateErrored:
		return "errored"
	}

	return "undefined"
}

// NumExecs returns number of registered WorkerProcess execs.
func (s *state) NumExecs() int64 {
	return atomic.LoadInt64(&s.numExecs)
}

// Value state returns state value
func (s *state) Value() int64 {
	return atomic.LoadInt64(&s.value)
}

// IsActive returns true if WorkerProcess not Inactive or Stopped
func (s *state) IsActive() bool {
	val := s.Value()
	return val == StateWorking || val == StateReady
}

// change state value (status)
func (s *state) Set(value int64) {
	atomic.StoreInt64(&s.value, value)
}

// register new execution atomically
func (s *state) RegisterExec() {
	atomic.AddInt64(&s.numExecs, 1)
}

// Update last used time
func (s *state) SetLastUsed(lu uint64) {
	atomic.StoreUint64(&s.lastUsed, lu)
}

func (s *state) LastUsed() uint64 {
	return atomic.LoadUint64(&s.lastUsed)
}
