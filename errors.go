package roadrunner

// JobError is job level error (no WorkerProcess halt), wraps at top
// of error context
type JobError []byte

// Error converts error context to string
func (te JobError) Error() string {
	return string(te)
}

// WorkerError is WorkerProcess related error
type WorkerError struct {
	// Worker
	Worker WorkerBase

	// Caused error
	Caused error
}

// Error converts error context to string
func (e WorkerError) Error() string {
	return e.Caused.Error()
}
