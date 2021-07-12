package ephemeral

import (
	"github.com/spiral/roadrunner/v2/common/jobs"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	PluginName string = "ephemeral"
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
	return PluginName
}

func (p *Plugin) Available() {}

// JobsConstruct creates new ephemeral consumer from the configuration
func (p *Plugin) JobsConstruct(configKey string, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return NewJobBroker(configKey, p.log, p.cfg, pq)
}

// FromPipeline creates new ephemeral consumer from the provided pipeline
func (p *Plugin) FromPipeline(pipeline *pipeline.Pipeline, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return FromPipeline(pipeline, p.log, pq)
}
