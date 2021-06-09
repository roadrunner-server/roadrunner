package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ChannelConfig configures loggers per channel.
type ChannelConfig struct {
	// Dedicated channels per logger. By default logger allocated via named logger.
	Channels map[string]Config `mapstructure:"channels"`
}

// FileLoggerConfig structure represents configuration for the file logger
type FileLoggerConfig struct {
	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.  It uses <processname>-lumberjack.log in
	// os.TempDir() if empty.
	LogOutput string `mapstructure:"log_output"`

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int `mapstructure:"max_size"`

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int `mapstructure:"max_age"`

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `mapstructure:"max_backups"`

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool `mapstructure:"compress"`
}

func (fl *FileLoggerConfig) InitDefaults() *FileLoggerConfig {
	if fl.LogOutput == "" {
		fl.LogOutput = os.TempDir()
	}

	if fl.MaxSize == 0 {
		fl.MaxSize = 100
	}

	if fl.MaxAge == 0 {
		fl.MaxAge = 24
	}

	if fl.MaxBackups == 0 {
		fl.MaxBackups = 10
	}

	return fl
}

type Config struct {
	// Mode configures logger based on some default template (development, production, off).
	Mode Mode `mapstructure:"mode"`

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

	// File logger options
	FileLogger *FileLoggerConfig `mapstructure:"file_logger_options"`
}

// BuildLogger converts config into Zap configuration.
func (cfg *Config) BuildLogger() (*zap.Logger, error) {
	var zCfg zap.Config
	switch Mode(strings.ToLower(string(cfg.Mode))) {
	case off, none:
		return zap.NewNop(), nil
	case production:
		zCfg = zap.NewProductionConfig()
	case development:
		zCfg = zap.Config{
			Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
			Development: true,
			Encoding:    "console",
			EncoderConfig: zapcore.EncoderConfig{
				// Keys can be anything except the empty string.
				TimeKey:        "T",
				LevelKey:       "L",
				NameKey:        "N",
				CallerKey:      "C",
				FunctionKey:    zapcore.OmitKey,
				MessageKey:     "M",
				StacktraceKey:  "S",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    ColoredLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
				EncodeName:     ColoredNameEncoder,
			},
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}
	case raw:
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

	// if we also have a file logger specified in the config
	// init it
	// otherwise - return standard config
	if cfg.FileLogger != nil {
		// init absent options
		cfg.FileLogger.InitDefaults()

		w := zapcore.AddSync(
			&lumberjack.Logger{
				Filename:   cfg.FileLogger.LogOutput,
				MaxSize:    cfg.FileLogger.MaxSize,
				MaxAge:     cfg.FileLogger.MaxAge,
				MaxBackups: cfg.FileLogger.MaxBackups,
				Compress:   cfg.FileLogger.Compress,
			},
		)

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(zCfg.EncoderConfig),
			w,
			zCfg.Level,
		)
		return zap.New(core), nil
	}

	return zCfg.Build()
}

// InitDefault Initialize default logger
func (cfg *Config) InitDefault() {
	if cfg.Mode == "" {
		cfg.Mode = development
	}
	if cfg.Level == "" {
		cfg.Level = "debug"
	}
}
