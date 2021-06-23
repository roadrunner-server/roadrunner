package ephemeral

import "github.com/spiral/roadrunner/v2/plugins/logger"

type Plugin struct {
	log logger.Logger
}

func (p *Plugin) Init(log logger.Logger) error {
	p.log = log
	return nil
}
