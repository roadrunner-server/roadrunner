package ephemeral

import (
	"github.com/google/uuid"
	"github.com/spiral/errors"
	priorityqueue "github.com/spiral/roadrunner/v2/common/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
)

type JobBroker struct {
	queues map[string]bool
	pq     priorityqueue.Queue
}

func NewJobBroker(q priorityqueue.Queue) (*JobBroker, error) {
	jb := &JobBroker{
		queues: make(map[string]bool),
		pq:     q,
	}

	return jb, nil
}

func (j *JobBroker) Push(job *structs.Job) (string, error) {
	const op = errors.Op("ephemeral_push")

	// check if the pipeline registered
	if b, ok := j.queues[job.Options.Pipeline]; ok {
		if !b {
			return "", errors.E(op, errors.Errorf("pipeline disabled: %s", job.Options.Pipeline))
		}
		if job.Options.Priority == nil {
			job.Options.Priority = intPtr(10)
		}
		job.Options.ID = uuid.NewString()

		j.pq.Insert(job)

		return job.Options.ID, nil
	}

	return "", errors.E(op, errors.Errorf("no such pipeline: %s", job.Options.Pipeline))
}

func (j *JobBroker) Stat() {
	panic("implement me")
}

func (j *JobBroker) Consume(pipe *pipeline.Pipeline) {
	panic("implement me")
}

func (j *JobBroker) Register(pipeline string) error {
	const op = errors.Op("ephemeral_register")
	if _, ok := j.queues[pipeline]; ok {
		return errors.E(op, errors.Errorf("queue %s has already been registered", pipeline))
	}

	j.queues[pipeline] = true

	return nil
}

func intPtr(val uint64) *uint64 {
	if val == 0 {
		val = 10
	}
	return &val
}
