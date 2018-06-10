package roadrunner

import (
	"fmt"
	"sync/atomic"
	"time"
)

// State represents worker status and updated time.
type State interface {
	fmt.Stringer

	// Value returns state value
	Value() int64

	// NumJobs shows how many times worker was invoked
	NumExecs() int64

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

	// StateStopping - process is being softly stopped.
	StateStopping

	// StateStopped - process has been terminated.
	StateStopped

	// StateErrored - error state (can't be used).
	StateErrored
)

type state struct {
	value    int64
	numExecs int64
	updated  int64
}

func newState(value int64) *state {
	return &state{value: value, updated: time.Now().Unix()}
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
	case StateStopped:
		return "stopped"
	case StateErrored:
		return "errored"
	}

	return "undefined"
}

// Value state returns state value
func (s *state) Value() int64 {
	return atomic.LoadInt64(&s.value)
}

// IsActive returns true if worker not Inactive or Stopped
func (s *state) IsActive() bool {
	state := s.Value()
	return state == StateWorking || state == StateReady
}

// Updated indicates a moment updated last state change
func (s *state) Updated() time.Time {
	return time.Unix(0, atomic.LoadInt64(&s.updated))
}

func (s *state) NumExecs() int64 {
	return atomic.LoadInt64(&s.numExecs)
}

// change state value (status)
func (s *state) set(value int64) {
	atomic.StoreInt64(&s.value, value)
	atomic.StoreInt64(&s.updated, time.Now().Unix())
}

// register new execution atomically
func (s *state) registerExec() {
	atomic.AddInt64(&s.numExecs, 1)
}
