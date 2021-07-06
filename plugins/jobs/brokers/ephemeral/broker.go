package ephemeral

import (
	"sync"

	"github.com/google/uuid"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/priorityqueue"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
	"github.com/spiral/roadrunner/v2/utils"
)

type JobBroker struct {
	queues sync.Map
	pq     priorityqueue.Queue
}

func NewJobBroker(q priorityqueue.Queue) (*JobBroker, error) {
	jb := &JobBroker{
		queues: sync.Map{},
		pq:     q,
	}

	return jb, nil
}

func (j *JobBroker) Push(job *structs.Job) (*string, error) {
	const op = errors.Op("ephemeral_push")

	// check if the pipeline registered
	if b, ok := j.queues.Load(job.Options.Pipeline); ok {
		if !b.(bool) {
			return nil, errors.E(op, errors.Errorf("pipeline disabled: %s", job.Options.Pipeline))
		}
		if job.Options.Priority == nil {
			job.Options.Priority = utils.AsUint64Ptr(10)
		}
		job.Options.ID = utils.AsStringPtr(uuid.NewString())

		j.pq.Insert(job)

		return job.Options.ID, nil
	}

	return nil, errors.E(op, errors.Errorf("no such pipeline: %s", job.Options.Pipeline))
}

func (j *JobBroker) Register(pipeline string) error {
	const op = errors.Op("ephemeral_register")
	if _, ok := j.queues.Load(pipeline); ok {
		return errors.E(op, errors.Errorf("queue %s has already been registered", pipeline))
	}

	j.queues.Store(pipeline, true)

	return nil
}

func (j *JobBroker) Stop(pipeline string) {
	if q, ok := j.queues.Load(pipeline); ok {
		if q == true {
			// mark pipeline as turned off
			j.queues.Store(pipeline, false)
		}
	}
}

func (j *JobBroker) StopAll() {
	j.queues.Range(func(key, value interface{}) bool {
		j.queues.Store(key, false)
		return true
	})
}

func (j *JobBroker) Resume(pipeline string) {
	if q, ok := j.queues.Load(pipeline); ok {
		if q == false {
			// mark pipeline as turned off
			j.queues.Store(pipeline, true)
		}
	}
}

func (j *JobBroker) ResumeAll() {
	j.queues.Range(func(key, value interface{}) bool {
		j.queues.Store(key, true)
		return true
	})
}

func (j *JobBroker) Stat() {
	panic("implement me")
}

func (j *JobBroker) Consume(pipe *pipeline.Pipeline) {
	panic("implement me")
}
