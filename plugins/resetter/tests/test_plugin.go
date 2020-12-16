package tests

import (
	"context"
	"time"

	"github.com/spiral/roadrunner/v2/interfaces/server"
	poolImpl "github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/plugins/config"
)

var testPoolConfig = poolImpl.Config{
	NumWorkers:      10,
	MaxJobs:         100,
	AllocateTimeout: time.Second * 10,
	DestroyTimeout:  time.Second * 10,
	Supervisor: &poolImpl.SupervisorConfig{
		WatchTick:       60,
		TTL:             1000,
		IdleTTL:         10,
		ExecTTL:         10,
		MaxWorkerMemory: 1000,
	},
}

// Gauge //////////////
type Plugin1 struct {
	config config.Configurer
	server server.Server
}

func (p1 *Plugin1) Init(cfg config.Configurer, server server.Server) error {
	p1.config = cfg
	p1.server = server
	return nil
}

func (p1 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p1 *Plugin1) Stop() error {
	return nil
}

func (p1 *Plugin1) Name() string {
	return "resetter.plugin1"
}

func (p1 *Plugin1) Reset() error {
	pool, err := p1.server.NewWorkerPool(context.Background(), testPoolConfig, nil)
	if err != nil {
		panic(err)
	}
	pool.Destroy(context.Background())

	pool, err = p1.server.NewWorkerPool(context.Background(), testPoolConfig, nil)
	if err != nil {
		panic(err)
	}

	_ = pool

	return nil
}
