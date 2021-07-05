package ephemeral

import (
	priorityqueue "github.com/spiral/roadrunner/v2/common/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/jobs"
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

func (p *Plugin) InitJobBroker(q priorityqueue.Queue) (jobs.Consumer, error) {
	return NewJobBroker(q)
}
