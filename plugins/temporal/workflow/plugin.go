package workflow

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/worker"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/spiral/roadrunner/v2/plugins/temporal/client"
)

const (
	// PluginName defines public service name.
	PluginName = "workflows"

	// RRMode sets as RR_MODE env variable to let worker know about the mode to run.
	RRMode = "temporal/workflow"
)

// Plugin manages workflows and workers.
type Plugin struct {
	temporal client.Temporal
	events   events.Handler
	server   server.Server
	log      logger.Logger
	mu       sync.Mutex
	reset    chan struct{}
	pool     workflowPool
	closing  int64
}

// Init workflow plugin.
func (p *Plugin) Init(temporal client.Temporal, server server.Server, log logger.Logger) error {
	p.temporal = temporal
	p.server = server
	p.events = events.NewEventsHandler()
	p.log = log
	p.reset = make(chan struct{}, 1)

	return nil
}

// Serve starts workflow service.
func (p *Plugin) Serve() chan error {
	const op = errors.Op("workflow_plugin_serve")
	errCh := make(chan error, 1)

	pool, err := p.startPool()
	if err != nil {
		errCh <- errors.E(op, err)
		return errCh
	}

	p.pool = pool

	go func() {
		for {
			select {
			case <-p.reset:
				if atomic.LoadInt64(&p.closing) == 1 {
					return
				}

				err := p.replacePool()
				if err == nil {
					continue
				}

				bkoff := backoff.NewExponentialBackOff()
				bkoff.InitialInterval = time.Second

				err = backoff.Retry(p.replacePool, bkoff)
				if err != nil {
					errCh <- errors.E(op, err)
				}
			}
		}
	}()

	return errCh
}

// Stop workflow service.
func (p *Plugin) Stop() error {
	const op = errors.Op("workflow_plugin_stop")
	atomic.StoreInt64(&p.closing, 1)

	pool := p.getPool()
	if pool != nil {
		p.pool = nil
		err := pool.Destroy(context.Background())
		if err != nil {
			return errors.E(op, err)
		}
		return nil
	}

	return nil
}

// Name of the service.
func (p *Plugin) Name() string {
	return PluginName
}

// Workers returns list of available workflow workers.
func (p *Plugin) Workers() []worker.BaseProcess {
	return p.pool.Workers()
}

// WorkflowNames returns list of all available workflows.
func (p *Plugin) WorkflowNames() []string {
	return p.pool.WorkflowNames()
}

// Reset resets underlying workflow pool with new copy.
func (p *Plugin) Reset() error {
	p.reset <- struct{}{}

	return nil
}

// AddListener adds event listeners to the service.
func (p *Plugin) poolListener(event interface{}) {
	if ev, ok := event.(PoolEvent); ok {
		if ev.Event == eventWorkerExit {
			if ev.Caused != nil {
				p.log.Error("Workflow pool error", "error", ev.Caused)
			}
			p.reset <- struct{}{}
		}
	}

	p.events.Push(event)
}

// AddListener adds event listeners to the service.
func (p *Plugin) startPool() (workflowPool, error) {
	const op = errors.Op("workflow_plugin_start_pool")
	pool, err := newWorkflowPool(
		p.temporal.GetCodec().WithLogger(p.log),
		p.poolListener,
		p.server,
	)
	if err != nil {
		return nil, errors.E(op, err)
	}

	err = pool.Start(context.Background(), p.temporal)
	if err != nil {
		return nil, errors.E(op, err)
	}

	p.log.Debug("Started workflow processing", "workflows", pool.WorkflowNames())

	return pool, nil
}

func (p *Plugin) replacePool() error {
	p.mu.Lock()
	const op = errors.Op("workflow_plugin_replace_pool")
	defer p.mu.Unlock()

	if p.pool != nil {
		err := p.pool.Destroy(context.Background())
		p.pool = nil
		if err != nil {
			p.log.Error(
				"Unable to destroy expired workflow pool",
				"error",
				errors.E(op, err),
			)
			return errors.E(op, err)
		}
	}

	pool, err := p.startPool()
	if err != nil {
		p.log.Error("Replace workflow pool failed", "error", err)
		return errors.E(op, err)
	}

	p.pool = pool
	p.log.Debug("workflow pool successfully replaced")

	return nil
}

// getPool returns currently pool.
func (p *Plugin) getPool() workflowPool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.pool
}
