package amqp

import (
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/common/jobs"
	"github.com/spiral/roadrunner/v2/pkg/priorityqueue"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/streadway/amqp"
)

type Config struct {
	Addr  string
	Queue string
}

type JobsConsumer struct {
	sync.RWMutex
	logger logger.Logger
	pq     priorityqueue.Queue

	pipelines sync.Map

	// amqp connection
	conn          *amqp.Connection
	retryTimeout  time.Duration
	prefetchCount int
	exchangeName  string
	connStr       string
	exchangeType  string
	routingKey    string

	stop chan struct{}
}

func NewAMQPConsumer(configKey string, log logger.Logger, cfg config.Configurer, pq priorityqueue.Queue) (jobs.Consumer, error) {
	// we need to obtain two parts of the amqp information here.
	// firs part - address to connect, it is located in the global section under the amqp name
	// second part - queues and other pipeline information
	jb := &JobsConsumer{
		logger: log,
		pq:     pq,
	}

	d, err := jb.initRabbitMQ()
	if err != nil {
		return nil, err
	}

	// run listener
	jb.listener(d)

	// run redialer
	jb.redialer()

	return jb, nil
}

func (j *JobsConsumer) Push(job *structs.Job) error {
	const op = errors.Op("ephemeral_push")

	// check if the pipeline registered
	if b, ok := j.pipelines.Load(job.Options.Pipeline); ok {
		if !b.(bool) {
			return errors.E(op, errors.Errorf("pipeline disabled: %s", job.Options.Pipeline))
		}

		// handle timeouts
		if job.Options.Timeout > 0 {
			go func(jj *structs.Job) {
				time.Sleep(jj.Options.TimeoutDuration())

				// TODO push

				// send the item after timeout expired
			}(job)

			return nil
		}

		// insert to the local, limited pipeline

		return nil
	}

	return errors.E(op, errors.Errorf("no such pipeline: %s", job.Options.Pipeline))
}

func (j *JobsConsumer) Register(pipeline *pipeline.Pipeline) error {
	panic("implement me")
}

func (j *JobsConsumer) List() []*pipeline.Pipeline {
	panic("implement me")
}

func (j *JobsConsumer) Pause(pipeline string) {
	panic("implement me")
}

func (j *JobsConsumer) Resume(pipeline string) {
	panic("implement me")
}
