package ephemeral

import (
	"github.com/spiral/roadrunner/v2/common/jobs"
	"github.com/spiral/roadrunner/v2/pkg/priorityqueue"
	"github.com/spiral/roadrunner/v2/plugins/config"
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

func (p *Plugin) JobsConstruct(configKey string, q priorityqueue.Queue) (jobs.Consumer, error) {
	return NewJobBroker(configKey, p.log, p.cfg, q)
}
