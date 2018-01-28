package roadrunner

import (
	"sync"
	"sync/atomic"
	"time"
)

// State represents worker status and updated time.
type State interface {
	// Value returns state value
	Value() int64

	// NumExecs shows how many times worker was invoked
	NumExecs() uint64

	// Updated indicates a moment updated last state change
	Updated() time.Time
}

const (
	// StateInactive - no associated process
	StateInactive int64 = iota
	// StateReady - ready for job.
	StateReady
	// StateWorking - working on given payload.
	StateWorking
	// StateStopped - process has been terminated
	StateStopped
	// StateErrored - error state (can't be used)
	StateErrored
)

type state struct {
	mu       sync.RWMutex
	value    int64
	numExecs uint64
	updated  time.Time
}

func newState(value int64) *state {
	return &state{value: value, updated: time.Now()}
}

// String returns current state as string.
func (s *state) String() string {
	switch s.value {
	case StateInactive:
		return "inactive"
	case StateReady:
		return "ready"
	case StateWorking:
		return "working"
	case StateStopped:
		return "stopped"
	case StateErrored:
		return "errored"
	}

	return "undefined"
}

// Value state returns state value
func (s *state) Value() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.value
}

// IsActive returns true if worker not Inactive or Stopped
func (s *state) IsActive() bool {
	state := s.Value()
	return state == StateWorking || state == StateReady
}

// Updated indicates a moment updated last state change
func (s *state) Updated() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.updated
}

func (s *state) NumExecs() uint64 {
	return atomic.LoadUint64(&s.numExecs)
}

// change state value (status)
func (s *state) set(value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.value = value
	s.updated = time.Now()
}

// register new execution atomically
func (s *state) registerExec() {
	atomic.AddUint64(&s.numExecs, 1)
}
