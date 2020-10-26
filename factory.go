package roadrunner

import (
	"context"
	"os/exec"
)

// Factory is responsible of wrapping given command into tasks WorkerProcess.
type Factory interface {
	// SpawnWorker creates new WorkerProcess process based on given command.
	// Process must not be started.
	SpawnWorkerWithContext(context.Context, *exec.Cmd) (WorkerBase, error)

	SpawnWorker(*exec.Cmd) (WorkerBase, error)

	// Close the factory and underlying connections.
	Close(ctx context.Context) error
}
