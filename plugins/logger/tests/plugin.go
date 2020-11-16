package tests

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/interfaces/log"
	"github.com/spiral/roadrunner/v2/plugins/config"
)

type Plugin struct {
	config config.Configurer
	log    log.Logger
}

func (p1 *Plugin) Init(cfg config.Configurer, log log.Logger) error {
	p1.config = cfg
	p1.log = log
	return nil
}

func (p1 *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	p1.log.Error("error", "test", errors.E(errors.Str("test")))
	p1.log.Info("error", "test", errors.E(errors.Str("test")))
	p1.log.Debug("error", "test", errors.E(errors.Str("test")))
	p1.log.Warn("error", "test", errors.E(errors.Str("test")))

	p1.log.Error("error", "test")
	p1.log.Info("error", "test")
	p1.log.Debug("error", "test")
	p1.log.Warn("error", "test")
	return errCh
}

func (p1 *Plugin) Stop() error {
	return nil
}

func (p1 *Plugin) Name() string {
	return "logger_plugin"
}
