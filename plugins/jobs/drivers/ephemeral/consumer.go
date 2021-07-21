package ephemeral

import (
	"sync"
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
	pipelineSize string = "pipeline_size"
)

type Config struct {
	PipelineSize uint64 `mapstructure:"pipeline_size"`
}

type JobBroker struct {
	cfg        *Config
	log        logger.Logger
	eh         events.Handler
	pipeline   sync.Map
	pq         priorityqueue.Queue
	localQueue chan *Item

	stopCh chan struct{}
}

func NewJobBroker(configKey string, log logger.Logger, cfg config.Configurer, eh events.Handler, pq priorityqueue.Queue) (*JobBroker, error) {
	const op = errors.Op("new_ephemeral_pipeline")

	jb := &JobBroker{
		log:    log,
		pq:     pq,
		eh:     eh,
		stopCh: make(chan struct{}, 1),
	}

	err := cfg.UnmarshalKey(configKey, &jb.cfg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if jb.cfg.PipelineSize == 0 {
		jb.cfg.PipelineSize = 100_000
	}

	// initialize a local queue
	jb.localQueue = make(chan *Item, jb.cfg.PipelineSize)

	// consume from the queue
	go jb.consume()

	return jb, nil
}

func FromPipeline(pipeline *pipeline.Pipeline, log logger.Logger, eh events.Handler, pq priorityqueue.Queue) (*JobBroker, error) {
	jb := &JobBroker{
		log:    log,
		pq:     pq,
		eh:     eh,
		stopCh: make(chan struct{}, 1),
	}

	jb.cfg.PipelineSize = uint64(pipeline.Int(pipelineSize, 100_000))

	// initialize a local queue
	jb.localQueue = make(chan *Item, jb.cfg.PipelineSize)

	// consume from the queue
	go jb.consume()

	return jb, nil
}

func (j *JobBroker) Push(jb *job.Job) error {
	const op = errors.Op("ephemeral_push")

	// check if the pipeline registered
	if b, ok := j.pipeline.Load(jb.Options.Pipeline); ok {
		if !b.(bool) {
			return errors.E(op, errors.Errorf("pipeline disabled: %s", jb.Options.Pipeline))
		}

		msg := fromJob(jb)
		// handle timeouts
		if msg.Options.Timeout > 0 {
			go func(jj *job.Job) {
				time.Sleep(jj.Options.TimeoutDuration())

				// send the item after timeout expired
				j.localQueue <- msg
			}(jb)

			return nil
		}

		// insert to the local, limited pipeline
		j.localQueue <- msg

		return nil
	}

	return errors.E(op, errors.Errorf("no such pipeline: %s", jb.Options.Pipeline))
}

func (j *JobBroker) consume() {
	// redirect
	for {
		select {
		case item := <-j.localQueue:
			j.pq.Insert(item)
		case <-j.stopCh:
			return
		}
	}
}

func (j *JobBroker) Register(pipeline *pipeline.Pipeline) error {
	const op = errors.Op("ephemeral_register")
	if _, ok := j.pipeline.Load(pipeline.Name()); ok {
		return errors.E(op, errors.Errorf("queue %s has already been registered", pipeline))
	}

	j.pipeline.Store(pipeline.Name(), true)

	return nil
}

func (j *JobBroker) Pause(pipeline string) {
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

func (j *JobBroker) Resume(pipeline string) {
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
func (j *JobBroker) Run(pipe *pipeline.Pipeline) error {
	j.eh.Push(events.JobEvent{
		Event:    events.EventPipeRun,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    time.Now(),
	})
	return nil
}

func (j *JobBroker) Stop() error {
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
