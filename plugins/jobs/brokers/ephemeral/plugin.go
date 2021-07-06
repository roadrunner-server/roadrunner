package ephemeral

import (
	"github.com/spiral/roadrunner/v2/common/jobs"
	"github.com/spiral/roadrunner/v2/pkg/priorityqueue"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	PluginName string = "ephemeral"
)

type Plugin struct {
	log logger.Logger
}

func (p *Plugin) Init(log logger.Logger) error {
	p.log = log
	return nil
}

func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) JobsConstruct(_ string, q priorityqueue.Queue) (jobs.Consumer, error) {
	return NewJobBroker(q)
}
