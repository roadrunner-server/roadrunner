package jobs

import (
	priorityqueue "github.com/spiral/roadrunner/v2/common/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
)

// Consumer todo naming
type Consumer interface {
	Push(*structs.Job) (string, error)
	Stat()
	Consume(*pipeline.Pipeline)
	Register(pipe string) error
}

type Broker interface {
	InitJobBroker(queue priorityqueue.Queue) (Consumer, error)
}

type Item interface {
	ID() string
	Payload() []byte
}
