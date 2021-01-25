package informer

import (
	"context"
	"time"

	"github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/pkg/worker"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/server"
)

var testPoolConfig = pool.Config{
	NumWorkers:      10,
	MaxJobs:         100,
	AllocateTimeout: time.Second * 10,
	DestroyTimeout:  time.Second * 10,
	Supervisor: &pool.SupervisorConfig{
		WatchTick:       60 * time.Second,
		TTL:             1000 * time.Second,
		IdleTTL:         10 * time.Second,
		ExecTTL:         10 * time.Second,
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
	return "informer.plugin1"
}

func (p1 *Plugin1) Workers() []worker.BaseProcess {
	p, err := p1.server.NewWorkerPool(context.Background(), testPoolConfig, nil)
	if err != nil {
		panic(err)
	}

	workers := p.Workers()
	baseWorkers := make([]worker.BaseProcess, 0, len(workers))
	for i := 0; i < len(workers); i++ {
		baseWorkers = append(baseWorkers, worker.FromSync(workers[i].(*worker.SyncWorkerImpl)))
	}

	return baseWorkers
}
