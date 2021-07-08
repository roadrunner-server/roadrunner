package amqp

import (
	"github.com/spiral/roadrunner/v2/common/jobs"
	"github.com/spiral/roadrunner/v2/pkg/priorityqueue"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type Config struct {
	Addr  string
	Queue string
}

type JobsConsumer struct {
	log logger.Logger
	pq  priorityqueue.Queue
}

func NewAMQPConsumer(configKey string, log logger.Logger, cfg config.Configurer, pq priorityqueue.Queue) (jobs.Consumer, error) {
	jb := &JobsConsumer{
		log: log,
		pq:  pq,
	}

	return jb, nil
}

func (j JobsConsumer) Push(job *structs.Job) error {
	panic("implement me")
}

func (j JobsConsumer) Register(pipeline *pipeline.Pipeline) error {
	panic("implement me")
}

func (j JobsConsumer) List() []*pipeline.Pipeline {
	panic("implement me")
}

func (j JobsConsumer) Pause(pipeline string) {
	panic("implement me")
}

func (j JobsConsumer) Resume(pipeline string) {
	panic("implement me")
}
