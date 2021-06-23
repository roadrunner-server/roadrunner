package ephemeral

import (
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
)

type JobBroker struct {
}

func NewJobBroker(q priorityqueue.Queue) (*JobBroker, error) {
	return &JobBroker{}, nil
}

func (j *JobBroker) Push(pipeline *pipeline.Pipeline, job *structs.Job) (string, error) {
	panic("implement me")
}

func (j *JobBroker) Stat() {
	panic("implement me")
}

func (j *JobBroker) Consume(pipeline *pipeline.Pipeline) {
	panic("implement me")
}

func (j *JobBroker) Register(pipeline *pipeline.Pipeline) {
	panic("implement me")
}
