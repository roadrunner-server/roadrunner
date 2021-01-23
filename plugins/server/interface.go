package server

import (
	"context"
	"os/exec"

	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/pkg/worker"
)

// Env variables type alias
type Env map[string]string

// Server creates workers for the application.
type Server interface {
	CmdFactory(env Env) (func() *exec.Cmd, error)
	NewWorker(ctx context.Context, env Env, listeners ...events.Listener) (*worker.Process, error)
	NewWorkerPool(ctx context.Context, opt pool.Config, env Env, listeners ...events.Listener) (pool.Pool, error)
}
