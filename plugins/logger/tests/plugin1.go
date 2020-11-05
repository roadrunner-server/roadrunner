package tests

import (
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type Plugin1 struct {
	config config.Configurer
	log    *logger.ZapLogger
}

func (p1 *Plugin1) Init(cfg config.Configurer, log *logger.ZapLogger) error {
	p1.config = cfg
	p1.log = log
	return nil
}

func (p1 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p1 *Plugin1) Stop() error {
	return nil
}
