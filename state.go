package roadrunner

// State is current state int.
type State int

const (
	// StateInactive - no associated process
	StateInactive State = iota
	// StateBooting - relay attached but w.Start() not executed
	StateBooting
	// StateReady - ready for job.
	StateReady
	// StateWorking - working on given payload.
	StateWorking
	// StateStopped - process has been terminated
	StateStopped
	// StateError - error State (can't be used)
	StateError
)

// String returns current state as string.
func (s State) String() string {
	switch s {
	case StateInactive:
		return "inactive"
	case StateBooting:
		return "booting"
	case StateReady:
		return "ready"
	case StateWorking:
		return "working"
	case StateStopped:
		return "stopped"
	case StateError:
		return "error"
	}

	return "undefined"
}
