package ephemeral

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	jobState "github.com/spiral/roadrunner/v2/pkg/state/job"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/job"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/utils"
)

const (
	prefetch      string = "prefetch"
	goroutinesMax uint64 = 1000
)

type Config struct {
	Prefetch uint64 `mapstructure:"prefetch"`
}

type consumer struct {
	cfg           *Config
	log           logger.Logger
	eh            events.Handler
	pipeline      atomic.Value
	pq            priorityqueue.Queue
	localPrefetch chan *Item

	// time.sleep goroutines max number
	goroutines uint64

	delayed *int64
	active  *int64

	listeners uint32
	stopCh    chan struct{}
}

func NewJobBroker(configKey string, log logger.Logger, cfg config.Configurer, eh events.Handler, pq priorityqueue.Queue) (*consumer, error) {
	const op = errors.Op("new_ephemeral_pipeline")

	jb := &consumer{
		log:        log,
		pq:         pq,
		eh:         eh,
		goroutines: 0,
		active:     utils.Int64(0),
		delayed:    utils.Int64(0),
		stopCh:     make(chan struct{}, 1),
	}

	err := cfg.UnmarshalKey(configKey, &jb.cfg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if jb.cfg.Prefetch == 0 {
		jb.cfg.Prefetch = 100_000
	}

	// initialize a local queue
	jb.localPrefetch = make(chan *Item, jb.cfg.Prefetch)

	return jb, nil
}

func FromPipeline(pipeline *pipeline.Pipeline, log logger.Logger, eh events.Handler, pq priorityqueue.Queue) (*consumer, error) {
	jb := &consumer{
		log:        log,
		pq:         pq,
		eh:         eh,
		goroutines: 0,
		active:     utils.Int64(0),
		delayed:    utils.Int64(0),
		stopCh:     make(chan struct{}, 1),
	}

	// initialize a local queue
	jb.localPrefetch = make(chan *Item, pipeline.Int(prefetch, 100_000))

	return jb, nil
}

func (j *consumer) Push(ctx context.Context, jb *job.Job) error {
	const op = errors.Op("ephemeral_push")

	// check if the pipeline registered
	_, ok := j.pipeline.Load().(*pipeline.Pipeline)
	if !ok {
		return errors.E(op, errors.Errorf("no such pipeline: %s", jb.Options.Pipeline))
	}

	err := j.handleItem(ctx, fromJob(jb))
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (j *consumer) State(_ context.Context) (*jobState.State, error) {
	pipe := j.pipeline.Load().(*pipeline.Pipeline)
	return &jobState.State{
		Pipeline: pipe.Name(),
		Driver:   pipe.Driver(),
		Queue:    pipe.Name(),
		Active:   atomic.LoadInt64(j.active),
		Delayed:  atomic.LoadInt64(j.delayed),
		Ready:    ready(atomic.LoadUint32(&j.listeners)),
	}, nil
}

func (j *consumer) Register(_ context.Context, pipeline *pipeline.Pipeline) error {
	j.pipeline.Store(pipeline)
	return nil
}

func (j *consumer) Pause(_ context.Context, p string) {
	pipe := j.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		j.log.Error("no such pipeline", "requested pause on: ", p)
	}

	l := atomic.LoadUint32(&j.listeners)
	// no active listeners
	if l == 0 {
		j.log.Warn("no active listeners, nothing to pause")
		return
	}

	atomic.AddUint32(&j.listeners, ^uint32(0))

	// stop the consumer
	j.stopCh <- struct{}{}

	j.eh.Push(events.JobEvent{
		Event:    events.EventPipePaused,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    time.Now(),
		Elapsed:  0,
	})
}

func (j *consumer) Resume(_ context.Context, p string) {
	pipe := j.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		j.log.Error("no such pipeline", "requested resume on: ", p)
	}

	l := atomic.LoadUint32(&j.listeners)
	// listener already active
	if l == 1 {
		j.log.Warn("listener already in the active state")
		return
	}

	// resume the consumer on the same channel
	j.consume()

	atomic.StoreUint32(&j.listeners, 1)
	j.eh.Push(events.JobEvent{
		Event:    events.EventPipeActive,
		Pipeline: pipe.Name(),
		Start:    time.Now(),
		Elapsed:  0,
	})
}

// Run is no-op for the ephemeral
func (j *consumer) Run(_ context.Context, pipe *pipeline.Pipeline) error {
	j.eh.Push(events.JobEvent{
		Event:    events.EventPipeActive,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    time.Now(),
	})
	return nil
}

func (j *consumer) Stop(ctx context.Context) error {
	const op = errors.Op("ephemeral_plugin_stop")

	pipe := j.pipeline.Load().(*pipeline.Pipeline)

	select {
	// return from the consumer
	case j.stopCh <- struct{}{}:
		j.eh.Push(events.JobEvent{
			Event:    events.EventPipeStopped,
			Pipeline: pipe.Name(),
			Start:    time.Now(),
			Elapsed:  0,
		})

		return nil

	case <-ctx.Done():
		return errors.E(op, ctx.Err())
	}
}

func (j *consumer) handleItem(ctx context.Context, msg *Item) error {
	const op = errors.Op("ephemeral_handle_request")
	// handle timeouts
	// theoretically, some bad user may send millions requests with a delay and produce a billion (for example)
	// goroutines here. We should limit goroutines here.
	if msg.Options.Delay > 0 {
		// if we have 1000 goroutines waiting on the delay - reject 1001
		if atomic.LoadUint64(&j.goroutines) >= goroutinesMax {
			return errors.E(op, errors.Str("max concurrency number reached"))
		}

		go func(jj *Item) {
			atomic.AddUint64(&j.goroutines, 1)
			atomic.AddInt64(j.delayed, 1)

			time.Sleep(jj.Options.DelayDuration())

			// send the item after timeout expired
			j.localPrefetch <- jj

			atomic.AddUint64(&j.goroutines, ^uint64(0))
		}(msg)

		return nil
	}

	// increase number of the active jobs
	atomic.AddInt64(j.active, 1)

	// insert to the local, limited pipeline
	select {
	case j.localPrefetch <- msg:
		return nil
	case <-ctx.Done():
		return errors.E(op, errors.Errorf("local pipeline is full, consider to increase prefetch number, current limit: %d, context error: %v", j.cfg.Prefetch, ctx.Err()))
	}
}

func (j *consumer) consume() {
	go func() {
		// redirect
		for {
			select {
			case item, ok := <-j.localPrefetch:
				if !ok {
					j.log.Warn("ephemeral local prefetch queue was closed")
					return
				}

				// set requeue channel
				item.Options.requeueFn = j.handleItem
				item.Options.active = j.active
				item.Options.delayed = j.delayed

				j.pq.Insert(item)
			case <-j.stopCh:
				return
			}
		}
	}()
}

func ready(r uint32) bool {
	return r > 0
}
