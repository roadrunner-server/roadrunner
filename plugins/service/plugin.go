package service

import (
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type Plugin struct {
}

func (p *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	return nil
}
