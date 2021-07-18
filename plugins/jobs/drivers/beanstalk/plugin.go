package beanstalk

import (
	"github.com/spiral/roadrunner/v2/common/jobs"
	"github.com/spiral/roadrunner/v2/pkg/events"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/logger"
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

func (p *Plugin) JobsConstruct(configKey string, eh events.Handler, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return nil, nil
}

func (p *Plugin) FromPipeline(pipe *pipeline.Pipeline, eh events.Handler, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return nil, nil
}
