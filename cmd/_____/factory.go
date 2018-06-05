package _____

import (
	"github.com/spiral/roadrunner"
	"net"
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

func (f *PoolConfig) NewServer() (*roadrunner.Server, func(), error) {
	relays, terminator, err := f.relayFactory()
	if err != nil {
		terminator()
		return nil, nil, err
	}

	rr := roadrunner.NewServer(f.cmd(), relays)
	if err := rr.Configure(f.rrConfig()); err != nil {
		return nil, nil, err
	}

	return rr, nil, nil
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

func (f *PoolConfig) relayFactory() (roadrunner.Factory, func(), error) {
	if f.Relay == "pipes" || f.Relay == "pipe" {
		return roadrunner.NewPipeFactory(), nil, nil
	}

	dsn := strings.Split(f.Relay, "://")
	if len(dsn) != 2 {
		return nil, nil, dsnError
	}

	ln, err := net.Listen(dsn[0], dsn[1])
	if err != nil {
		return nil, nil, err
	}

	return roadrunner.NewSocketFactory(ln, time.Minute), func() { ln.Close() }, nil
}
