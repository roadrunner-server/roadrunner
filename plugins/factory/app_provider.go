package factory

import (
	"os/exec"

	"github.com/temporalio/roadrunner-temporal/roadrunner"
)

type Env map[string]string

type Spawner interface {
	// CmdFactory create new command factory with given env variables.
	NewCmd(env Env) (func() *exec.Cmd, error)

	// NewFactory inits new factory for workers.
	NewFactory(env Env) (roadrunner.Factory, error)
}
