package memory

import (
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const PluginName string = "memory"

type Plugin struct {
	log logger.Logger
}

func (p *Plugin) Init(log logger.Logger) error {
	p.log = log
	return nil
}

func (p *Plugin) PSProvide(key string) (pubsub.PubSub, error) {
	return NewPubSubDriver(p.log, key)
}

func (p *Plugin) Name() string {
	return PluginName
}
