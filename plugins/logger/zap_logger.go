package logger

import (
	"github.com/spiral/endure"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"go.uber.org/zap"
)

// ServiceName declares service name.
const ServiceName = "logs"

type LogFactory interface {
	// GlobalLogger returns global log instance.
	GlobalLogger() *zap.Logger

	// NamedLogger returns logger dedicated to the specific channel. Similar to Named() but also reads the core params.
	NamedLogger(name string) *zap.Logger
}

// ZapLogger manages zap logger.
type ZapLogger struct {
	base     *zap.Logger
	cfg      Config
	channels ChannelConfig
}

// Init logger service.
func (z *ZapLogger) Init(cfg config.Provider) (err error) {
	err = cfg.UnmarshalKey(ServiceName, &z.cfg)
	if err != nil {
		return err
	}

	err = cfg.UnmarshalKey(ServiceName, &z.channels)
	if err != nil {
		return err
	}

	z.base, err = z.cfg.BuildLogger()
	return err
}

// DefaultLogger returns default logger.
func (z *ZapLogger) DefaultLogger() (*zap.Logger, error) {
	return z.base, nil
}

// NamedLogger returns logger dedicated to the specific channel. Similar to Named() but also reads the core params.
func (z *ZapLogger) NamedLogger(name string) (*zap.Logger, error) {
	if cfg, ok := z.channels.Channels[name]; ok {
		return cfg.BuildLogger()
	}

	return z.base.Named(name), nil
}

// NamedLogger returns logger dedicated to the specific channel. Similar to Named() but also reads the core params.
func (z *ZapLogger) ServiceLogger(n endure.Named) (*zap.Logger, error) {
	return z.NamedLogger(n.Name())
}

// Provides declares factory methods.
func (z *ZapLogger) Provides() []interface{} {
	return []interface{}{
		z.DefaultLogger,
		z.ServiceLogger,
	}
}
