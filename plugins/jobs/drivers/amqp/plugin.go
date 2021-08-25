package amqp

import (
	"github.com/spiral/roadrunner/v2/common/jobs"
	"github.com/spiral/roadrunner/v2/pkg/events"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/drivers/amqp/amqpjobs"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	pluginName string = "amqp"
)

type Plugin struct {
	log logger.Logger
	cfg config.Configurer
}

func (p *Plugin) Init(log logger.Logger, cfg config.Configurer) error {
	p.log = log
	p.cfg = cfg
	return nil
}

func (p *Plugin) Name() string {
	return pluginName
}

func (p *Plugin) Available() {}

func (p *Plugin) JobsConstruct(configKey string, e events.Handler, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return amqpjobs.NewAMQPConsumer(configKey, p.log, p.cfg, e, pq)
}

// FromPipeline constructs AMQP driver from pipeline
func (p *Plugin) FromPipeline(pipe *pipeline.Pipeline, e events.Handler, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return amqpjobs.FromPipeline(pipe, p.log, p.cfg, e, pq)
}
