package logger

import (
	"strings"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type Plugin struct {
	config config.Configurer
	log    logger.Logger
}

func (p1 *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
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

	// test the `raw` mode
	messageJSON := []byte(`{"field": "value"}`)
	p1.log.Debug(strings.TrimRight(string(messageJSON), " \n\t"))

	return errCh
}

func (p1 *Plugin) Stop() error {
	return nil
}

func (p1 *Plugin) Name() string {
	return "logger_plugin"
}
