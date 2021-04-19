package worker

import (
	"sync/atomic"
)

// SYNC WITH worker_watcher.GET
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

	// StateKilling - process is being forcibly stopped
	StateKilling

	// StateDestroyed State of worker, when no need to allocate new one
	StateDestroyed

	// StateMaxJobsReached State of worker, when it reached executions limit
	StateMaxJobsReached

	// StateStopped - process has been terminated.
	StateStopped

	// StateErrored - error StateImpl (can't be used).
	StateErrored
)

type StateImpl struct {
	value    int64
	numExecs uint64
	// to be lightweight, use UnixNano
	lastUsed uint64
}

// Thread safe
func NewWorkerState(value int64) *StateImpl {
	return &StateImpl{value: value}
}

// String returns current StateImpl as string.
func (s *StateImpl) String() string {
	switch s.Value() {
	case StateInactive:
		return "inactive"
	case StateReady:
		return "ready"
	case StateWorking:
		return "working"
	case StateInvalid:
		return "invalid"
	case StateStopping:
		return "stopping"
	case StateStopped:
		return "stopped"
	case StateKilling:
		return "killing"
	case StateErrored:
		return "errored"
	case StateDestroyed:
		return "destroyed"
	}

	return "undefined"
}

// NumExecs returns number of registered WorkerProcess execs.
func (s *StateImpl) NumExecs() uint64 {
	return atomic.LoadUint64(&s.numExecs)
}

// Value StateImpl returns StateImpl value
func (s *StateImpl) Value() int64 {
	return atomic.LoadInt64(&s.value)
}

// IsActive returns true if WorkerProcess not Inactive or Stopped
func (s *StateImpl) IsActive() bool {
	val := s.Value()
	return val == StateWorking || val == StateReady
}

// Set change StateImpl value (status)
func (s *StateImpl) Set(value int64) {
	atomic.StoreInt64(&s.value, value)
}

// RegisterExec register new execution atomically
func (s *StateImpl) RegisterExec() {
	atomic.AddUint64(&s.numExecs, 1)
}

// SetLastUsed Update last used time
func (s *StateImpl) SetLastUsed(lu uint64) {
	atomic.StoreUint64(&s.lastUsed, lu)
}

func (s *StateImpl) LastUsed() uint64 {
	return atomic.LoadUint64(&s.lastUsed)
}
