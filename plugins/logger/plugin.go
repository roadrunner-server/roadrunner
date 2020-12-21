package logger

import (
	"github.com/spiral/endure"
	"github.com/spiral/roadrunner/v2/interfaces/config"
	"github.com/spiral/roadrunner/v2/interfaces/log"
	"go.uber.org/zap"
)

// PluginName declares plugin name.
const PluginName = "logs"

// ZapLogger manages zap logger.
type ZapLogger struct {
	base     *zap.Logger
	cfg      Config
	channels ChannelConfig
}

// Init logger service.
func (z *ZapLogger) Init(cfg config.Configurer) error {
	err := cfg.UnmarshalKey(PluginName, &z.cfg)
	if err != nil {
		return err
	}

	err = cfg.UnmarshalKey(PluginName, &z.channels)
	if err != nil {
		return err
	}

	z.base, err = z.cfg.BuildLogger()
	return err
}

// DefaultLogger returns default logger.
func (z *ZapLogger) DefaultLogger() (log.Logger, error) {
	return NewZapAdapter(z.base), nil
}

// NamedLogger returns logger dedicated to the specific channel. Similar to Named() but also reads the core params.
func (z *ZapLogger) NamedLogger(name string) (log.Logger, error) {
	if cfg, ok := z.channels.Channels[name]; ok {
		l, err := cfg.BuildLogger()
		if err != nil {
			return nil, err
		}
		return NewZapAdapter(l), nil
	}

	return NewZapAdapter(z.base.Named(name)), nil
}

// NamedLogger returns logger dedicated to the specific channel. Similar to Named() but also reads the core params.
func (z *ZapLogger) ServiceLogger(n endure.Named) (log.Logger, error) {
	return z.NamedLogger(n.Name())
}

// Provides declares factory methods.
func (z *ZapLogger) Provides() []interface{} {
	return []interface{}{
		z.ServiceLogger,
		z.DefaultLogger,
	}
}
