package jobs

import (
	"github.com/spiral/roadrunner/v2/pkg/priorityqueue"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
)

// Consumer todo naming
type Consumer interface {
	Push(job *structs.Job) (*string, error)
	PushBatch(job *[]structs.Job) (*string, error)
	Consume(job *pipeline.Pipeline)

	Stop(pipeline string)
	StopAll()
	Resume(pipeline string)
	ResumeAll()

	Register(pipe string) error
	Stat()
}

type Constructor interface {
	JobsConstruct(configKey string, queue priorityqueue.Queue) (Consumer, error)
}
