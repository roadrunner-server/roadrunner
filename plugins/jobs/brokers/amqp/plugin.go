package amqp

import (
	"github.com/spiral/roadrunner/v2/common/jobs"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/config"
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

func (p *Plugin) JobsConstruct(configKey string, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return NewAMQPConsumer(configKey, p.log, p.cfg, pq)
}

// FromPipeline constructs AMQP driver from pipeline
func (p *Plugin) FromPipeline(pipe *pipeline.Pipeline, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return FromPipeline(pipe, pq)
}
