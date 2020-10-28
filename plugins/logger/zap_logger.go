package logger

import (
	"github.com/spiral/endure"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	base *zap.Logger
	cfg  Config
}

func (z *ZapLogger) Init(cfg config.Provider) (err error) {
	err = cfg.UnmarshalKey(ServiceName, &z.cfg)
	if err != nil {
		return err
	}

	if z.base == nil {
		cfg := zap.Config{
			Level:    zap.NewAtomicLevelAt(zap.DebugLevel),
			Encoding: "console",
			EncoderConfig: zapcore.EncoderConfig{
				MessageKey:   "message",
				LevelKey:     "level",
				TimeKey:      "time",
				NameKey:      "name",
				EncodeName:   ColoredHashedNameEncoder,
				EncodeLevel:  ColoredLevelEncoder,
				EncodeTime:   UTCTimeEncoder,
				EncodeCaller: zapcore.ShortCallerEncoder,
			},
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}

		z.base, err = cfg.Build()
		if err != nil {
			return err
		}
	}

	return nil
}

// GlobalLogger returns global log instance.
func (z *ZapLogger) GlobalLogger() *zap.Logger {
	return z.base
}

// NamedLogger returns logger dedicated to the specific channel. Similar to Named() but also reads the core params.
func (z *ZapLogger) NamedLogger(name string) *zap.Logger {
	// todo: automatically configure
	return z.base.Named(name)
}

// Provides declares factory methods.
func (z *ZapLogger) Provides() []interface{} {
	return []interface{}{
		z.DefaultLogger,
		z.AllocateLogger,
	}
}

// AllocateLogger allocates logger for the service.
func (z *ZapLogger) AllocateLogger(n endure.Named) (*zap.Logger, error) {
	return z.NamedLogger(n.Name()), nil
}

// DefaultLogger returns default logger.
func (z *ZapLogger) DefaultLogger() (*zap.Logger, error) {
	return z.GlobalLogger(), nil
}
