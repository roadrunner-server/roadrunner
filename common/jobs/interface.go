package jobs

import (
	"context"

	"github.com/spiral/roadrunner/v2/pkg/events"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	jobState "github.com/spiral/roadrunner/v2/pkg/state/job"
	"github.com/spiral/roadrunner/v2/plugins/jobs/job"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
)

// Consumer represents a single jobs driver interface
type Consumer interface {
	Push(ctx context.Context, job *job.Job) error
	Register(ctx context.Context, pipeline *pipeline.Pipeline) error
	Run(ctx context.Context, pipeline *pipeline.Pipeline) error
	Stop(ctx context.Context) error

	Pause(ctx context.Context, pipeline string)
	Resume(ctx context.Context, pipeline string)

	// State provide information about driver state
	State(ctx context.Context) (*jobState.State, error)
}

// Constructor constructs Consumer interface. Endure abstraction.
type Constructor interface {
	JobsConstruct(configKey string, e events.Handler, queue priorityqueue.Queue) (Consumer, error)
	FromPipeline(pipe *pipeline.Pipeline, e events.Handler, queue priorityqueue.Queue) (Consumer, error)
}
