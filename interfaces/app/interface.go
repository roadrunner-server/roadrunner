package app

import (
	"context"
	"os/exec"

	"github.com/spiral/roadrunner/v2"
)

// WorkerFactory creates workers for the application.
type WorkerFactory interface {
	CmdFactory(env map[string]string) (func() *exec.Cmd, error)
	NewWorker(ctx context.Context, env map[string]string) (roadrunner.WorkerBase, error)
	NewWorkerPool(ctx context.Context, opt roadrunner.PoolConfig, env map[string]string) (roadrunner.Pool, error)
}
