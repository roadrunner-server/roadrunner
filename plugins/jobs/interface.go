package jobs

import (
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
)

// Consumer todo naming
type Consumer interface {
	Push(*pipeline.Pipeline, *structs.Job) (string, error)
	Stat()
	Consume(*pipeline.Pipeline)
	Register(*pipeline.Pipeline)
}

type Broker interface {
	InitJobBroker(queue priorityqueue.Queue) (Consumer, error)
}
