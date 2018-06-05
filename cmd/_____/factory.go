package _____

import (
	"github.com/spiral/roadrunner"
	"os/exec"
	"strings"
	"time"
)

// todo: move out
type PoolConfig struct {
	Command string
	Relay   string

	Number  uint64
	MaxJobs uint64

	Timeouts struct {
		Construct int
		Allocate  int
		Destroy   int
	}
}

func (f *PoolConfig) rrConfig() roadrunner.Config {
	return roadrunner.Config{
		NumWorkers:      f.Number,
		MaxExecutions:   f.MaxJobs,
		AllocateTimeout: time.Second * time.Duration(f.Timeouts.Allocate),
		DestroyTimeout:  time.Second * time.Duration(f.Timeouts.Destroy),
	}
}

func (f *PoolConfig) cmd() func() *exec.Cmd {
	cmd := strings.Split(f.Command, " ")
	return func() *exec.Cmd { return exec.Command(cmd[0], cmd[1:]...) }
}
