package worker

import (
	"context"
	"os/exec"

	"github.com/spiral/roadrunner/v2/interfaces/worker"
)

// Factory is responsible of wrapping given command into tasks WorkerProcess.
type Factory interface {
	// SpawnWorkerWithContext creates new WorkerProcess process based on given command with contex.
	// Process must not be started.
	SpawnWorkerWithContext(context.Context, *exec.Cmd) (worker.BaseProcess, error)

	// SpawnWorker creates new WorkerProcess process based on given command.
	// Process must not be started.
	SpawnWorker(*exec.Cmd) (worker.BaseProcess, error)

	// Close the factory and underlying connections.
	Close(ctx context.Context) error
}
