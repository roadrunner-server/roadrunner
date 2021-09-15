package grpc

import (
	"context"
	"sync"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/pkg/state/process"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/grpc/codec"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/spiral/roadrunner/v2/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

const (
	name   string = "grpc"
	RrGrpc string = "RR_GRPC"
)

type Plugin struct {
	mu       *sync.RWMutex
	config   *Config
	gPool    pool.Pool
	opts     []grpc.ServerOption
	services []func(server *grpc.Server)
	server   *grpc.Server
	rrServer server.Server

	// events handler
	events events.Handler
	log    logger.Logger
}

func (p *Plugin) Init(cfg config.Configurer, log logger.Logger, server server.Server) error {
	const op = errors.Op("grpc_plugin_init")

	if !cfg.Has(name) {
		return errors.E(errors.Disabled)
	}
	// register the codec
	encoding.RegisterCodec(&codec.Codec{})

	err := cfg.UnmarshalKey(name, &p.config)
	if err != nil {
		return errors.E(op, err)
	}

	err = p.config.InitDefaults()
	if err != nil {
		return errors.E(op, err)
	}

	p.opts = make([]grpc.ServerOption, 0)
	p.services = make([]func(server *grpc.Server), 0)
	p.events = events.NewEventsHandler()
	p.events.AddListener(p.collectGRPCEvents)
	p.rrServer = server

	// worker's GRPC mode
	if p.config.Env == nil {
		p.config.Env = make(map[string]string)
	}
	p.config.Env[RrGrpc] = "true"

	p.log = log
	p.mu = &sync.RWMutex{}

	return nil
}

func (p *Plugin) Serve() chan error {
	const op = errors.Op("grpc_plugin_serve")
	errCh := make(chan error, 1)

	var err error
	p.gPool, err = p.rrServer.NewWorkerPool(context.Background(), &pool.Config{
		Debug:           p.config.GrpcPool.Debug,
		NumWorkers:      p.config.GrpcPool.NumWorkers,
		MaxJobs:         p.config.GrpcPool.MaxJobs,
		AllocateTimeout: p.config.GrpcPool.AllocateTimeout,
		DestroyTimeout:  p.config.GrpcPool.DestroyTimeout,
		Supervisor:      p.config.GrpcPool.Supervisor,
	}, p.config.Env, p.collectGRPCEvents)
	if err != nil {
		errCh <- errors.E(op, err)
		return errCh
	}

	go func() {
		var err error
		p.mu.Lock()
		p.server, err = p.createGRPCserver()
		if err != nil {
			p.log.Error("create grpc server", "error", err)
			errCh <- errors.E(op, err)
			return
		}

		l, err := utils.CreateListener(p.config.Listen)
		if err != nil {
			p.log.Error("create grpc listener", "error", err)
			errCh <- errors.E(op, err)
		}

		// protect serve
		p.mu.Unlock()
		err = p.server.Serve(l)
		if err != nil {
			// skip errors when stopping the server
			if err == grpc.ErrServerStopped {
				return
			}

			p.log.Error("grpc server stopped", "error", err)
			errCh <- errors.E(op, err)
			return
		}
	}()

	return errCh
}

func (p *Plugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.server != nil {
		p.server.Stop()
	}
	return nil
}

func (p *Plugin) Available() {}

func (p *Plugin) Name() string {
	return name
}

func (p *Plugin) Reset() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	const op = errors.Op("grpc_plugin_reset")

	// destroy old pool
	p.gPool.Destroy(context.Background())

	var err error
	p.gPool, err = p.rrServer.NewWorkerPool(context.Background(), &pool.Config{
		Debug:           p.config.GrpcPool.Debug,
		NumWorkers:      p.config.GrpcPool.NumWorkers,
		MaxJobs:         p.config.GrpcPool.MaxJobs,
		AllocateTimeout: p.config.GrpcPool.AllocateTimeout,
		DestroyTimeout:  p.config.GrpcPool.DestroyTimeout,
		Supervisor:      p.config.GrpcPool.Supervisor,
	}, p.config.Env, p.collectGRPCEvents)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (p *Plugin) Workers() []*process.State {
	p.mu.RLock()
	defer p.mu.RUnlock()

	workers := p.gPool.Workers()

	ps := make([]*process.State, 0, len(workers))
	for i := 0; i < len(workers); i++ {
		state, err := process.WorkerProcessState(workers[i])
		if err != nil {
			return nil
		}
		ps = append(ps, state)
	}

	return ps
}

func (p *Plugin) collectGRPCEvents(event interface{}) {
	if gev, ok := event.(events.GRPCEvent); ok {
		switch gev.Event {
		case events.EventUnaryCallOk:
			p.log.Info("method called", "method", gev.Info.FullMethod, "started", gev.Start, "elapsed", gev.Elapsed)
		case events.EventUnaryCallErr:
			p.log.Info("method call finished with error", "error", gev.Error, "method", gev.Info.FullMethod, "started", gev.Start, "elapsed", gev.Elapsed)
		}
	}
}
