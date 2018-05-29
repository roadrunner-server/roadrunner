package roadrunner

// JobError is job level error (no worker halt), wraps at top
// of error context
type JobError []byte

// Error converts error context to string
func (je JobError) Error() string {
	return string(je)
}

// WorkerError is worker related error
type WorkerError struct {
	// Worker
	Worker *Worker

	// Caused error
	Caused error
}

// Error converts error context to string
func (e WorkerError) Error() string {
	return e.Caused.Error()
}
