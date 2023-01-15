package container

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

type Config struct {
	GracePeriod time.Duration
	PrintGraph  bool
	LogLevel    slog.Leveler
}

const (
	endureKey          = "endure"
	defaultGracePeriod = time.Second * 30
)

// NewConfig creates endure container configuration.
func NewConfig(cfgFile string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(cfgFile)

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	if !v.IsSet(endureKey) {
		return &Config{ // return config with defaults
			GracePeriod: defaultGracePeriod,
			PrintGraph:  false,
			LogLevel:    slog.LevelError,
		}, nil
	}

	rrCfgEndure := struct {
		GracePeriod time.Duration `mapstructure:"grace_period"`
		PrintGraph  bool          `mapstructure:"print_graph"`
		LogLevel    string        `mapstructure:"log_level"`
	}{}

	err = v.UnmarshalKey(endureKey, &rrCfgEndure)
	if err != nil {
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
		LogLevel:    logLevel,
	}, nil
}

func parseLogLevel(s string) (slog.Leveler, error) {
	switch s {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	}

	return slog.LevelInfo, fmt.Errorf(`unknown log level "%s" (allowed: debug, info, warn, error, panic, fatal)`, s)
}
