package container

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

// Config defines endure container configuration.
type Config struct {
	GracePeriod time.Duration `mapstructure:"grace_period"`
	LogLevel    string        `mapstructure:"log_level"`
	WatchdogSec int           `mapstructure:"watchdog_sec"`
	PrintGraph  bool          `mapstructure:"print_graph"`
}

const (
	// endure config key
	endureKey = "endure"
	// overall grace period, after which container will be stopped forcefully
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

	cfg := &Config{
		GracePeriod: defaultGracePeriod,
		LogLevel:    "error",
		PrintGraph:  false,
	}

	if !v.IsSet(endureKey) {
		return cfg, nil
	}

	err = v.UnmarshalKey(endureKey, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func ParseLogLevel(s string) (slog.Leveler, error) {
	switch s {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelError, fmt.Errorf(`unknown log level "%s" (allowed: debug, info, warn, error)`, s)
	}
}
