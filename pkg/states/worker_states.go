package states

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

	// State of worker, when no need to allocate new one
	StateDestroyed

	// StateStopped - process has been terminated.
	StateStopped

	// StateErrored - error WorkerState (can't be used).
	StateErrored

	// StateRemove - worker is killed and removed from the stack
	StateRemove
)
