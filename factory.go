package roadrunner

import "os/exec"

// Factory is responsible of wrapping given command into tasks worker.
type Factory interface {
	// NewWorker creates new worker process based on given process.
	NewWorker(cmd *exec.Cmd) (w *Worker, err error)

	// Close closes all open factory descriptors.
	Close() error
}
