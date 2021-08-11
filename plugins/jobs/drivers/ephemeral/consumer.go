package ephemeral

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/job"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	prefetch string = "prefetch"
)

type Config struct {
	Prefetch uint64 `mapstructure:"prefetch"`
}

type JobConsumer struct {
	cfg           *Config
	log           logger.Logger
	eh            events.Handler
	pipeline      sync.Map
	pq            priorityqueue.Queue
	localPrefetch chan *Item

	// time.sleep goroutines max number
	goroutinesMaxNum uint64

	requeueCh chan *Item
	stopCh    chan struct{}
}

func NewJobBroker(configKey string, log logger.Logger, cfg config.Configurer, eh events.Handler, pq priorityqueue.Queue) (*JobConsumer, error) {
	const op = errors.Op("new_ephemeral_pipeline")

	jb := &JobConsumer{
		log:              log,
		pq:               pq,
		eh:               eh,
		goroutinesMaxNum: 1000,
		stopCh:           make(chan struct{}, 1),
		requeueCh:        make(chan *Item, 1000),
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

	// consume from the queue
	go jb.consume()
	jb.requeueListener()

	return jb, nil
}

func FromPipeline(pipeline *pipeline.Pipeline, log logger.Logger, eh events.Handler, pq priorityqueue.Queue) (*JobConsumer, error) {
	jb := &JobConsumer{
		log:              log,
		pq:               pq,
		eh:               eh,
		goroutinesMaxNum: 1000,
		stopCh:           make(chan struct{}, 1),
		requeueCh:        make(chan *Item, 1000),
	}

	// initialize a local queue
	jb.localPrefetch = make(chan *Item, pipeline.Int(prefetch, 100_000))

	// consume from the queue
	go jb.consume()
	jb.requeueListener()

	return jb, nil
}

func (j *JobConsumer) Push(ctx context.Context, jb *job.Job) error {
	const op = errors.Op("ephemeral_push")

	// check if the pipeline registered
	b, ok := j.pipeline.Load(jb.Options.Pipeline)
	if !ok {
		return errors.E(op, errors.Errorf("no such pipeline: %s", jb.Options.Pipeline))
	}

	if !b.(bool) {
		return errors.E(op, errors.Errorf("pipeline disabled: %s", jb.Options.Pipeline))
	}

	err := j.handleItem(ctx, fromJob(jb))
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (j *JobConsumer) handleItem(ctx context.Context, msg *Item) error {
	const op = errors.Op("ephemeral_handle_request")
	// handle timeouts
	// theoretically, some bad user may send millions requests with a delay and produce a billion (for example)
	// goroutines here. We should limit goroutines here.
	if msg.Options.Delay > 0 {
		// if we have 1000 goroutines waiting on the delay - reject 1001
		if atomic.LoadUint64(&j.goroutinesMaxNum) >= 1000 {
			return errors.E(op, errors.Str("max concurrency number reached"))
		}

		go func(jj *Item) {
			atomic.AddUint64(&j.goroutinesMaxNum, 1)
			time.Sleep(jj.Options.DelayDuration())

			// send the item after timeout expired
			j.localPrefetch <- jj

			atomic.AddUint64(&j.goroutinesMaxNum, ^uint64(0))
		}(msg)

		return nil
	}

	// insert to the local, limited pipeline
	select {
	case j.localPrefetch <- msg:
		return nil
	case <-ctx.Done():
		return errors.E(op, errors.Errorf("local pipeline is full, consider to increase prefetch number, current limit: %d, context error: %v", j.cfg.Prefetch, ctx.Err()))
	}
}

func (j *JobConsumer) consume() {
	// redirect
	for {
		select {
		case item := <-j.localPrefetch:

			// set requeue channel
			item.Options.requeueCh = j.requeueCh
			j.pq.Insert(item)
		case <-j.stopCh:
			return
		}
	}
}

func (j *JobConsumer) Register(_ context.Context, pipeline *pipeline.Pipeline) error {
	const op = errors.Op("ephemeral_register")
	if _, ok := j.pipeline.Load(pipeline.Name()); ok {
		return errors.E(op, errors.Errorf("queue %s has already been registered", pipeline))
	}

	j.pipeline.Store(pipeline.Name(), true)

	return nil
}

func (j *JobConsumer) Pause(_ context.Context, pipeline string) {
	if q, ok := j.pipeline.Load(pipeline); ok {
		if q == true {
			// mark pipeline as turned off
			j.pipeline.Store(pipeline, false)
		}
		// if not true - do not send the EventPipeStopped, because pipe already stopped
		return
	}

	j.eh.Push(events.JobEvent{
		Event:    events.EventPipePaused,
		Pipeline: pipeline,
		Start:    time.Now(),
		Elapsed:  0,
	})
}

func (j *JobConsumer) Resume(_ context.Context, pipeline string) {
	if q, ok := j.pipeline.Load(pipeline); ok {
		if q == false {
			// mark pipeline as turned on
			j.pipeline.Store(pipeline, true)
		}

		// if not true - do not send the EventPipeActive, because pipe already active
		return
	}

	j.eh.Push(events.JobEvent{
		Event:    events.EventPipeActive,
		Pipeline: pipeline,
		Start:    time.Now(),
		Elapsed:  0,
	})
}

// Run is no-op for the ephemeral
func (j *JobConsumer) Run(_ context.Context, pipe *pipeline.Pipeline) error {
	j.eh.Push(events.JobEvent{
		Event:    events.EventPipeActive,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    time.Now(),
	})
	return nil
}

func (j *JobConsumer) Stop(ctx context.Context) error {
	const op = errors.Op("ephemeral_plugin_stop")
	var pipe string
	j.pipeline.Range(func(key, _ interface{}) bool {
		pipe = key.(string)
		j.pipeline.Delete(key)
		return true
	})

	select {
	// return from the consumer
	case j.stopCh <- struct{}{}:
		j.eh.Push(events.JobEvent{
			Event:    events.EventPipeStopped,
			Pipeline: pipe,
			Start:    time.Now(),
			Elapsed:  0,
		})
		return nil

	case <-ctx.Done():
		return errors.E(op, ctx.Err())
	}
}
