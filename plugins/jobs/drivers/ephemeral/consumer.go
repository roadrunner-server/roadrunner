package ephemeral

import (
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

	goroutinesMaxNum uint64

	stopCh chan struct{}
}

func NewJobBroker(configKey string, log logger.Logger, cfg config.Configurer, eh events.Handler, pq priorityqueue.Queue) (*JobConsumer, error) {
	const op = errors.Op("new_ephemeral_pipeline")

	jb := &JobConsumer{
		log:              log,
		pq:               pq,
		eh:               eh,
		goroutinesMaxNum: 1000,
		stopCh:           make(chan struct{}, 1),
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

	return jb, nil
}

func FromPipeline(pipeline *pipeline.Pipeline, log logger.Logger, eh events.Handler, pq priorityqueue.Queue) (*JobConsumer, error) {
	jb := &JobConsumer{
		log:              log,
		pq:               pq,
		eh:               eh,
		goroutinesMaxNum: 1000,
		stopCh:           make(chan struct{}, 1),
	}

	// initialize a local queue
	jb.localPrefetch = make(chan *Item, pipeline.Int(prefetch, 100_000))

	// consume from the queue
	go jb.consume()

	return jb, nil
}

func (j *JobConsumer) Push(jb *job.Job) error {
	const op = errors.Op("ephemeral_push")

	// check if the pipeline registered
	if b, ok := j.pipeline.Load(jb.Options.Pipeline); ok {
		if !b.(bool) {
			return errors.E(op, errors.Errorf("pipeline disabled: %s", jb.Options.Pipeline))
		}

		msg := fromJob(jb)
		// handle timeouts
		// theoretically, some bad user may send a millions requests with a delay and produce a billion (for example)
		// goroutines here. We should limit goroutines here.
		if msg.Options.Delay > 0 {
			// if we have 1000 goroutines waiting on the delay - reject 1001
			if atomic.LoadUint64(&j.goroutinesMaxNum) >= 1000 {
				return errors.E(op, errors.Str("max concurrency number reached"))
			}

			go func(jj *job.Job) {
				atomic.AddUint64(&j.goroutinesMaxNum, 1)
				time.Sleep(jj.Options.DelayDuration())

				// send the item after timeout expired
				j.localPrefetch <- msg

				atomic.AddUint64(&j.goroutinesMaxNum, ^uint64(0))
			}(jb)

			return nil
		}

		// insert to the local, limited pipeline
		select {
		case j.localPrefetch <- msg:
		default:
			return errors.E(op, errors.Errorf("local pipeline is full, consider to increase prefetch number, current limit: %d", j.cfg.Prefetch))
		}

		return nil
	}

	return errors.E(op, errors.Errorf("no such pipeline: %s", jb.Options.Pipeline))
}

func (j *JobConsumer) consume() {
	// redirect
	for {
		select {
		case item := <-j.localPrefetch:
			j.pq.Insert(item)
		case <-j.stopCh:
			return
		}
	}
}

func (j *JobConsumer) Register(pipeline *pipeline.Pipeline) error {
	const op = errors.Op("ephemeral_register")
	if _, ok := j.pipeline.Load(pipeline.Name()); ok {
		return errors.E(op, errors.Errorf("queue %s has already been registered", pipeline))
	}

	j.pipeline.Store(pipeline.Name(), true)

	return nil
}

func (j *JobConsumer) Pause(pipeline string) {
	if q, ok := j.pipeline.Load(pipeline); ok {
		if q == true {
			// mark pipeline as turned off
			j.pipeline.Store(pipeline, false)
		}
	}

	j.eh.Push(events.JobEvent{
		Event:    events.EventPipeStopped,
		Pipeline: pipeline,
		Start:    time.Now(),
		Elapsed:  0,
	})
}

func (j *JobConsumer) Resume(pipeline string) {
	if q, ok := j.pipeline.Load(pipeline); ok {
		if q == false {
			// mark pipeline as turned on
			j.pipeline.Store(pipeline, true)
		}
	}

	j.eh.Push(events.JobEvent{
		Event:    events.EventPipeActive,
		Pipeline: pipeline,
		Start:    time.Now(),
		Elapsed:  0,
	})
}

// Run is no-op for the ephemeral
func (j *JobConsumer) Run(pipe *pipeline.Pipeline) error {
	j.eh.Push(events.JobEvent{
		Event:    events.EventPipeActive,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    time.Now(),
	})
	return nil
}

func (j *JobConsumer) Stop() error {
	var pipe string
	j.pipeline.Range(func(key, _ interface{}) bool {
		pipe = key.(string)
		j.pipeline.Delete(key)
		return true
	})

	// return from the consumer
	j.stopCh <- struct{}{}

	j.eh.Push(events.JobEvent{
		Event:    events.EventPipeStopped,
		Pipeline: pipe,
		Start:    time.Now(),
		Elapsed:  0,
	})

	return nil
}
