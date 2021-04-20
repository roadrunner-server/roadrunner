package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ChannelConfig configures loggers per channel.
type ChannelConfig struct {
	// Dedicated channels per logger. By default logger allocated via named logger.
	Channels map[string]Config `mapstructure:"channels"`
}

type Config struct {
	// Mode configures logger based on some default template (development, production, off).
	Mode string `mapstructure:"mode"`

	// Level is the minimum enabled logging level. Note that this is a dynamic
	// level, so calling ChannelConfig.Level.SetLevel will atomically change the log
	// level of all loggers descended from this config.
	Level string `mapstructure:"level"`

	// Encoding sets the logger's encoding. InitDefault values are "json" and
	// "console", as well as any third-party encodings registered via
	// RegisterEncoder.
	Encoding string `mapstructure:"encoding"`

	// Output is a list of URLs or file paths to write logging output to.
	// See Open for details.
	Output []string `mapstructure:"output"`

	// ErrorOutput is a list of URLs to write internal logger errors to.
	// The default is standard error.
	//
	// Note that this setting only affects internal errors; for sample code that
	// sends error-level logs to a different location from info- and debug-level
	// logs, see the package-level AdvancedConfiguration example.
	ErrorOutput []string `mapstructure:"errorOutput"`
}

// ZapConfig converts config into Zap configuration.
func (cfg *Config) BuildLogger() (*zap.Logger, error) {
	var zCfg zap.Config
	switch strings.ToLower(cfg.Mode) {
	case "off", "none":
		return zap.NewNop(), nil
	case "production":
		zCfg = zap.NewProductionConfig()
	case "development":
		zCfg = zap.NewDevelopmentConfig()
	case "raw":
		zCfg = zap.Config{
			Level:    zap.NewAtomicLevelAt(zap.InfoLevel),
			Encoding: "console",
			EncoderConfig: zapcore.EncoderConfig{
				MessageKey: "message",
			},
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}
	default:
		zCfg = zap.Config{
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
	}

	if cfg.Level != "" {
		level := zap.NewAtomicLevel()
		if err := level.UnmarshalText([]byte(cfg.Level)); err == nil {
			zCfg.Level = level
		}
	}

	if cfg.Encoding != "" {
		zCfg.Encoding = cfg.Encoding
	}

	if len(cfg.Output) != 0 {
		zCfg.OutputPaths = cfg.Output
	}

	if len(cfg.ErrorOutput) != 0 {
		zCfg.ErrorOutputPaths = cfg.ErrorOutput
	}

	// todo:

	return zCfg.Build()
}

// Initialize default logger
func (cfg *Config) InitDefault() {
	if cfg.Mode == "" {
		cfg.Mode = "development"
	}
	if cfg.Level == "" {
		cfg.Level = "debug"
	}
}
