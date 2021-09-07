package logger

import (
	"strings"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"go.uber.org/zap"
	core "go.uber.org/zap/zapcore"
)

type Plugin struct {
	config config.Configurer
	log    logger.Logger
}

type Loggable struct {
}

func (l *Loggable) MarshalLogObject(encoder core.ObjectEncoder) error {
	encoder.AddString("error", "Example marshaller error")
	return nil
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

	field := zap.String("error", "Example field error")

	p1.log.Error("error", field)
	p1.log.Info("error", field)
	p1.log.Debug("error", field)
	p1.log.Warn("error", field)

	marshalledObject := &Loggable{}

	p1.log.Error("error", marshalledObject)
	p1.log.Info("error", marshalledObject)
	p1.log.Debug("error", marshalledObject)
	p1.log.Warn("error", marshalledObject)

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
