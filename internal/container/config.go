package container

import (
	"fmt"
	"time"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner-plugins/v2/config"
)

type Config struct {
	GracePeriod time.Duration
	PrintGraph  bool
	RetryOnFail bool // TODO check for races, disabled at this moment
	LogLevel    endure.Level
}

const (
	endureKey          = "endure"
	defaultGracePeriod = time.Second * 30
)

// NewConfig creates endure container configuration.
func NewConfig(cfgPlugin *config.Plugin) (*Config, error) {
	if !cfgPlugin.Has(endureKey) {
		return &Config{ // return config with defaults
			GracePeriod: defaultGracePeriod,
			PrintGraph:  false,
			RetryOnFail: false,
			LogLevel:    endure.ErrorLevel,
		}, nil
	}

	rrCfgEndure := struct {
		GracePeriod time.Duration `mapstructure:"grace_period"`
		PrintGraph  bool          `mapstructure:"print_graph"`
		RetryOnFail bool          `mapstructure:"retry_on_fail"`
		LogLevel    string        `mapstructure:"log_level"`
	}{}

	if err := cfgPlugin.UnmarshalKey(endureKey, &rrCfgEndure); err != nil {
		return nil, err
	}

	if rrCfgEndure.GracePeriod == 0 {
		rrCfgEndure.GracePeriod = defaultGracePeriod
	}

	if rrCfgEndure.LogLevel == "" {
		rrCfgEndure.LogLevel = "error"
	}

	logLevel, err := parseLogLevel(rrCfgEndure.LogLevel)
	if err != nil {
		return nil, err
	}

	return &Config{
		GracePeriod: rrCfgEndure.GracePeriod,
		PrintGraph:  rrCfgEndure.PrintGraph,
		RetryOnFail: rrCfgEndure.RetryOnFail,
		LogLevel:    logLevel,
	}, nil
}

func parseLogLevel(s string) (endure.Level, error) {
	switch s {
	case "debug":
		return endure.DebugLevel, nil
	case "info":
		return endure.InfoLevel, nil
	case "warn", "warning":
		return endure.WarnLevel, nil
	case "error":
		return endure.ErrorLevel, nil
	case "panic":
		return endure.PanicLevel, nil
	case "fatal":
		return endure.FatalLevel, nil
	}

	return endure.DebugLevel, fmt.Errorf(`unknown log level "%s" (allowed: debug, info, warn, error, panic, fatal)`, s)
}
