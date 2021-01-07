package server

import (
	"context"
	"os/exec"

	"github.com/spiral/roadrunner/v2/interfaces/events"
	"github.com/spiral/roadrunner/v2/interfaces/pool"
	"github.com/spiral/roadrunner/v2/interfaces/worker"
	poolImpl "github.com/spiral/roadrunner/v2/pkg/pool"
)

// Env variables type alias
type Env map[string]string

// Server creates workers for the application.
type Server interface {
	CmdFactory(env Env) (func() *exec.Cmd, error)
	NewWorker(ctx context.Context, env Env, listeners ...events.Listener) (worker.BaseProcess, error)
	NewWorkerPool(ctx context.Context, opt poolImpl.Config, env Env, listeners ...events.Listener) (pool.Pool, error)
}
