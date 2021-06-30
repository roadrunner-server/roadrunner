package ephemeral

import (
	"github.com/google/uuid"
	"github.com/spiral/errors"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
)

type JobBroker struct {
	jobs   chan *entry
	queues map[*pipeline.Pipeline]*queue
	pq     priorityqueue.Queue
}

func NewJobBroker(q priorityqueue.Queue) (*JobBroker, error) {
	jb := &JobBroker{
		jobs: make(chan *entry, 10),
		pq:   q,
	}

	go jb.serve()

	return jb, nil
}

func (j *JobBroker) Push(pipe *pipeline.Pipeline, job *structs.Job) (string, error) {
	id := uuid.NewString()

	j.jobs <- &entry{
		id: id,
	}

	return id, nil
}

func (j *JobBroker) Stat() {
	panic("implement me")
}

func (j *JobBroker) Consume(pipeline *pipeline.Pipeline) {
	panic("implement me")
}

func (j *JobBroker) Register(pipeline *pipeline.Pipeline) error {
	const op = errors.Op("ephemeral_register")
	if _, ok := j.queues[pipeline]; !ok {
		return errors.E(op, errors.Errorf("queue %s has already been registered", pipeline.Name()))
	}

	j.queues[pipeline] = newQueue()

	return nil
}

func (j *JobBroker) serve() {
	for item := range j.jobs {
		// item should satisfy
		j.pq.Push(item)
	}
}
