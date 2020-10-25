package factory

import (
	"os/exec"

	"github.com/spiral/roadrunner/v2"
)

type Env map[string]string

type Spawner interface {
	// CmdFactory create new command factory with given env variables.
	NewCmd(env Env) (func() *exec.Cmd, error)

	// NewFactory inits new factory for workers.
	NewFactory() (roadrunner.Factory, error)
}
