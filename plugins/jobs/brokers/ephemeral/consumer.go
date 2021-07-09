package ephemeral

import (
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/priorityqueue"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type Config struct {
	PipelineSize uint64 `mapstructure:"pipeline_size"`
}

type JobBroker struct {
	cfg        *Config
	log        logger.Logger
	queues     sync.Map
	pq         priorityqueue.Queue
	localQueue chan *Item
}

func NewJobBroker(configKey string, log logger.Logger, cfg config.Configurer, q priorityqueue.Queue) (*JobBroker, error) {
	const op = errors.Op("new_ephemeral_pipeline")

	jb := &JobBroker{
		log: log,
		pq:  q,
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

func (j *JobBroker) Push(job *structs.Job) error {
	const op = errors.Op("ephemeral_push")

	// check if the pipeline registered
	if b, ok := j.queues.Load(job.Options.Pipeline); ok {
		if !b.(bool) {
			return errors.E(op, errors.Errorf("pipeline disabled: %s", job.Options.Pipeline))
		}

		// handle timeouts
		if job.Options.Timeout > 0 {
			go func(jj *structs.Job) {
				time.Sleep(jj.Options.TimeoutDuration())

				// send the item after timeout expired
				j.localQueue <- From(job)
			}(job)

			return nil
		}

		// insert to the local, limited pipeline
		j.localQueue <- From(job)

		return nil
	}

	return errors.E(op, errors.Errorf("no such pipeline: %s", job.Options.Pipeline))
}

func (j *JobBroker) consume() {
	// redirect
	for item := range j.localQueue {
		j.pq.Insert(item)
	}
}

func (j *JobBroker) Register(pipeline *pipeline.Pipeline) error {
	const op = errors.Op("ephemeral_register")
	if _, ok := j.queues.Load(pipeline.Name()); ok {
		return errors.E(op, errors.Errorf("queue %s has already been registered", pipeline))
	}

	j.queues.Store(pipeline.Name(), true)

	return nil
}

// Consume is no-op for the ephemeral
func (j *JobBroker) Consume(_ *pipeline.Pipeline) error {
	return nil
}

func (j *JobBroker) Pause(pipeline string) {
	if q, ok := j.queues.Load(pipeline); ok {
		if q == true {
			// mark pipeline as turned off
			j.queues.Store(pipeline, false)
		}
	}
}

func (j *JobBroker) Resume(pipeline string) {
	if q, ok := j.queues.Load(pipeline); ok {
		if q == false {
			// mark pipeline as turned off
			j.queues.Store(pipeline, true)
		}
	}
}

func (j *JobBroker) List() []*pipeline.Pipeline {
	out := make([]*pipeline.Pipeline, 0, 2)

	j.queues.Range(func(key, value interface{}) bool {
		pipe := key.(*pipeline.Pipeline)
		out = append(out, pipe)
		return true
	})

	return out
}
