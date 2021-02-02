package logger

import (
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"go.uber.org/zap"
)

// PluginName declares plugin name.
const PluginName = "logs"

// ZapLogger manages zap logger.
type ZapLogger struct {
	base     *zap.Logger
	cfg      *Config
	channels ChannelConfig
}

// Init logger service.
func (z *ZapLogger) Init(cfg config.Configurer) error {
	const op = errors.Op("config_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &z.cfg)
	if err != nil {
		return errors.E(op, errors.Disabled, err)
	}

	err = cfg.UnmarshalKey(PluginName, &z.channels)
	if err != nil {
		return errors.E(op, errors.Disabled, err)
	}

	z.base, err = z.cfg.BuildLogger()
	if err != nil {
		return errors.E(op, errors.Disabled, err)
	}
	return nil
}

// DefaultLogger returns default logger.
func (z *ZapLogger) DefaultLogger() (Logger, error) {
	return NewZapAdapter(z.base), nil
}

// NamedLogger returns logger dedicated to the specific channel. Similar to Named() but also reads the core params.
func (z *ZapLogger) NamedLogger(name string) (Logger, error) {
	if cfg, ok := z.channels.Channels[name]; ok {
		l, err := cfg.BuildLogger()
		if err != nil {
			return nil, err
		}
		return NewZapAdapter(l.Named(name)), nil
	}

	return NewZapAdapter(z.base.Named(name)), nil
}

// NamedLogger returns logger dedicated to the specific channel. Similar to Named() but also reads the core params.
func (z *ZapLogger) ServiceLogger(n endure.Named) (Logger, error) {
	return z.NamedLogger(n.Name())
}

// Provides declares factory methods.
func (z *ZapLogger) Provides() []interface{} {
	return []interface{}{
		z.ServiceLogger,
	}
}
