package jobs

import (
	"github.com/spiral/roadrunner/v2/pkg/priorityqueue"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
)

// Consumer todo naming
type Consumer interface {
	Push(job *structs.Job) error
	Register(pipeline *pipeline.Pipeline) error
	List() []*pipeline.Pipeline

	Pause(pipeline string)
	Resume(pipeline string)
}

type Constructor interface {
	JobsConstruct(configKey string, queue priorityqueue.Queue) (Consumer, error)
}
