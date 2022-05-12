package container

import (
	"fmt"
	"time"

	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/spf13/viper"
)

type Config struct {
	GracePeriod time.Duration
	PrintGraph  bool
	LogLevel    endure.Level
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
			LogLevel:    endure.ErrorLevel,
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
