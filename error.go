package roadrunner

// WorkerError is communication/process error.
type WorkerError string

// Error converts error context to string
func (we WorkerError) Error() string {
	return string(we)
}

// JobError is job level error (no worker halt)
type JobError []byte

// Error converts error context to string
func (je JobError) Error() string {
	return string(je)
}
