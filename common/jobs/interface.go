package jobs

import (
	"github.com/spiral/roadrunner/v2/pkg/events"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/jobs/job"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
)

// Consumer todo naming
type Consumer interface {
	Push(job *job.Job) error
	Register(pipeline *pipeline.Pipeline) error
	Run(pipeline *pipeline.Pipeline) error
	Stop() error

	Pause(pipeline string)
	Resume(pipeline string)
}

type Constructor interface {
	JobsConstruct(configKey string, e events.Handler, queue priorityqueue.Queue) (Consumer, error)
	FromPipeline(pipe *pipeline.Pipeline, e events.Handler, queue priorityqueue.Queue) (Consumer, error)
}
