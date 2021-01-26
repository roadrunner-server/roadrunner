package workflow

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/payload"
	rrWorker "github.com/spiral/roadrunner/v2/pkg/worker"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/spiral/roadrunner/v2/plugins/temporal/client"
	rrt "github.com/spiral/roadrunner/v2/plugins/temporal/protocol"
	bindings "go.temporal.io/sdk/internalbindings"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

const eventWorkerExit = 8390

// RR_MODE env variable key
const RR_MODE = "RR_MODE" //nolint

// RR_CODEC env variable key
const RR_CODEC = "RR_CODEC" //nolint

type workflowPool interface {
	SeqID() uint64
	Exec(p payload.Payload) (payload.Payload, error)
	Start(ctx context.Context, temporal client.Temporal) error
	Destroy(ctx context.Context) error
	Workers() []rrWorker.BaseProcess
	WorkflowNames() []string
}

// PoolEvent triggered on workflow pool worker events.
type PoolEvent struct {
	Event   int
	Context interface{}
	Caused  error
}

// workflowPoolImpl manages workflowProcess executions between worker restarts.
type workflowPoolImpl struct {
	codec     rrt.Codec
	seqID     uint64
	workflows map[string]rrt.WorkflowInfo
	tWorkers  []worker.Worker
	mu        sync.Mutex
	worker    rrWorker.SyncWorker
	active    bool
}

// newWorkflowPool creates new workflow pool.
func newWorkflowPool(codec rrt.Codec, listener events.Listener, factory server.Server) (workflowPool, error) {
	const op = errors.Op("new_workflow_pool")
	w, err := factory.NewWorker(
		context.Background(),
		map[string]string{RR_MODE: RRMode, RR_CODEC: codec.GetName()},
		listener,
	)
	if err != nil {
		return nil, errors.E(op, err)
	}

	go func() {
		err := w.Wait()
		listener(PoolEvent{Event: eventWorkerExit, Caused: err})
	}()

	return &workflowPoolImpl{codec: codec, worker: rrWorker.From(w)}, nil
}

// Start the pool in non blocking mode.
func (pool *workflowPoolImpl) Start(ctx context.Context, temporal client.Temporal) error {
	const op = errors.Op("workflow_pool_start")
	pool.mu.Lock()
	pool.active = true
	pool.mu.Unlock()

	err := pool.initWorkers(ctx, temporal)
	if err != nil {
		return errors.E(op, err)
	}

	for i := 0; i < len(pool.tWorkers); i++ {
		err := pool.tWorkers[i].Start()
		if err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}

// Active.
func (pool *workflowPoolImpl) Active() bool {
	return pool.active
}

// Destroy stops all temporal workers and application worker.
func (pool *workflowPoolImpl) Destroy(ctx context.Context) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	const op = errors.Op("workflow_pool_destroy")

	pool.active = false
	for i := 0; i < len(pool.tWorkers); i++ {
		pool.tWorkers[i].Stop()
	}

	worker.PurgeStickyWorkflowCache()

	err := pool.worker.Stop()
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

// NewWorkflowDefinition initiates new workflow process.
func (pool *workflowPoolImpl) NewWorkflowDefinition() bindings.WorkflowDefinition {
	return &workflowProcess{
		codec: pool.codec,
		pool:  pool,
	}
}

// NewWorkflowDefinition initiates new workflow process.
func (pool *workflowPoolImpl) SeqID() uint64 {
	return atomic.AddUint64(&pool.seqID, 1)
}

// Exec set of commands in thread safe move.
func (pool *workflowPoolImpl) Exec(p payload.Payload) (payload.Payload, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if !pool.active {
		return payload.Payload{}, nil
	}

	return pool.worker.Exec(p)
}

func (pool *workflowPoolImpl) Workers() []rrWorker.BaseProcess {
	return []rrWorker.BaseProcess{pool.worker}
}

func (pool *workflowPoolImpl) WorkflowNames() []string {
	names := make([]string, 0, len(pool.workflows))
	for name := range pool.workflows {
		names = append(names, name)
	}

	return names
}

// initWorkers request workers workflows from underlying PHP and configures temporal workers linked to the pool.
func (pool *workflowPoolImpl) initWorkers(ctx context.Context, temporal client.Temporal) error {
	const op = errors.Op("workflow_pool_init_workers")
	workerInfo, err := rrt.FetchWorkerInfo(pool.codec, pool, temporal.GetDataConverter())
	if err != nil {
		return errors.E(op, err)
	}

	pool.workflows = make(map[string]rrt.WorkflowInfo)
	pool.tWorkers = make([]worker.Worker, 0)

	for _, info := range workerInfo {
		w, err := temporal.CreateWorker(info.TaskQueue, info.Options)
		if err != nil {
			return errors.E(op, err, pool.Destroy(ctx))
		}

		pool.tWorkers = append(pool.tWorkers, w)
		for _, workflowInfo := range info.Workflows {
			w.RegisterWorkflowWithOptions(pool, workflow.RegisterOptions{
				Name:                          workflowInfo.Name,
				DisableAlreadyRegisteredCheck: false,
			})

			pool.workflows[workflowInfo.Name] = workflowInfo
		}
	}

	return nil
}
