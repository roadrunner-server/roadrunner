package activity

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/pool"
	rrWorker "github.com/spiral/roadrunner/v2/pkg/worker"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/spiral/roadrunner/v2/plugins/temporal/client"
	rrt "github.com/spiral/roadrunner/v2/plugins/temporal/protocol"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/internalbindings"
	"go.temporal.io/sdk/worker"
)

// RR_MODE env variable
const RR_MODE = "RR_MODE" //nolint:golint,stylecheck
// RR_CODEC env variable
const RR_CODEC = "RR_CODEC" //nolint:golint,stylecheck

//
const doNotCompleteOnReturn = "doNotCompleteOnReturn"

type activityPool interface {
	Start(ctx context.Context, temporal client.Temporal) error
	Destroy(ctx context.Context) error
	Workers() []rrWorker.SyncWorker
	ActivityNames() []string
	GetActivityContext(taskToken []byte) (context.Context, error)
}

type activityPoolImpl struct {
	dc         converter.DataConverter
	codec      rrt.Codec
	seqID      uint64
	activities []string
	wp         pool.Pool
	tWorkers   []worker.Worker
	running    sync.Map
}

// newActivityPool
func newActivityPool(codec rrt.Codec, listener events.Listener, poolConfig pool.Config, server server.Server) (activityPool, error) {
	const op = errors.Op("new_activity_pool")
	// env variables
	env := map[string]string{RR_MODE: RRMode, RR_CODEC: codec.GetName()}
	wp, err := server.NewWorkerPool(context.Background(), poolConfig, env, listener)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return &activityPoolImpl{
		codec:   codec,
		wp:      wp,
		running: sync.Map{},
	}, nil
}

// initWorkers request workers info from underlying PHP and configures temporal workers linked to the pool.
func (pool *activityPoolImpl) Start(ctx context.Context, temporal client.Temporal) error {
	const op = errors.Op("activity_pool_start")
	pool.dc = temporal.GetDataConverter()

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

// initWorkers request workers info from underlying PHP and configures temporal workers linked to the pool.
func (pool *activityPoolImpl) Destroy(ctx context.Context) error {
	for i := 0; i < len(pool.tWorkers); i++ {
		pool.tWorkers[i].Stop()
	}

	pool.wp.Destroy(ctx)
	return nil
}

// Workers returns list of all allocated workers.
func (pool *activityPoolImpl) Workers() []rrWorker.SyncWorker {
	return pool.wp.Workers()
}

// ActivityNames returns list of all available activity names.
func (pool *activityPoolImpl) ActivityNames() []string {
	return pool.activities
}

// ActivityNames returns list of all available activity names.
func (pool *activityPoolImpl) GetActivityContext(taskToken []byte) (context.Context, error) {
	const op = errors.Op("activity_pool_get_activity_context")
	c, ok := pool.running.Load(string(taskToken))
	if !ok {
		return nil, errors.E(op, errors.Str("heartbeat on non running activity"))
	}

	return c.(context.Context), nil
}

// initWorkers request workers workflows from underlying PHP and configures temporal workers linked to the pool.
func (pool *activityPoolImpl) initWorkers(ctx context.Context, temporal client.Temporal) error {
	const op = errors.Op("activity_pool_create_temporal_worker")

	workerInfo, err := rrt.FetchWorkerInfo(pool.codec, pool.wp, temporal.GetDataConverter())
	if err != nil {
		return errors.E(op, err)
	}

	pool.activities = make([]string, 0)
	pool.tWorkers = make([]worker.Worker, 0)

	for i := 0; i < len(workerInfo); i++ {
		w, err := temporal.CreateWorker(workerInfo[i].TaskQueue, workerInfo[i].Options)
		if err != nil {
			return errors.E(op, err, pool.Destroy(ctx))
		}

		pool.tWorkers = append(pool.tWorkers, w)
		for j := 0; j < len(workerInfo[i].Activities); j++ {
			w.RegisterActivityWithOptions(pool.executeActivity, activity.RegisterOptions{
				Name:                          workerInfo[i].Activities[j].Name,
				DisableAlreadyRegisteredCheck: false,
			})

			pool.activities = append(pool.activities, workerInfo[i].Activities[j].Name)
		}
	}

	return nil
}

// executes activity with underlying worker.
func (pool *activityPoolImpl) executeActivity(ctx context.Context, args *common.Payloads) (*common.Payloads, error) {
	const op = errors.Op("activity_pool_execute_activity")

	heartbeatDetails := &common.Payloads{}
	if activity.HasHeartbeatDetails(ctx) {
		err := activity.GetHeartbeatDetails(ctx, &heartbeatDetails)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}

	var info = activity.GetInfo(ctx)
	var msg = rrt.Message{
		ID: atomic.AddUint64(&pool.seqID, 1),
		Command: rrt.InvokeActivity{
			Name:             info.ActivityType.Name,
			Info:             info,
			HeartbeatDetails: len(heartbeatDetails.Payloads),
		},
		Payloads: args,
	}

	if len(heartbeatDetails.Payloads) != 0 {
		msg.Payloads.Payloads = append(msg.Payloads.Payloads, heartbeatDetails.Payloads...)
	}

	pool.running.Store(string(info.TaskToken), ctx)
	defer pool.running.Delete(string(info.TaskToken))

	result, err := pool.codec.Execute(pool.wp, rrt.Context{TaskQueue: info.TaskQueue}, msg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if len(result) != 1 {
		return nil, errors.E(op, errors.Str("invalid activity worker response"))
	}

	out := result[0]
	if out.Failure != nil {
		if out.Failure.Message == doNotCompleteOnReturn {
			return nil, activity.ErrResultPending
		}

		return nil, internalbindings.ConvertFailureToError(out.Failure, pool.dc)
	}

	return out.Payloads, nil
}
