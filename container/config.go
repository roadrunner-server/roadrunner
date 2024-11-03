package container

import (
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

// envVarPattern matches ${var} or $var or ${var:-default} patterns
var envVarPattern = regexp.MustCompile(`\$\{([^{}:\-]+)(?::-([^{}]+))?\}|\$([A-Za-z0-9_]+)`)

// expandEnvVars replaces environment variables in the input string
func expandEnvVars(input string) string {
	return envVarPattern.ReplaceAllStringFunc(input, func(match string) string {
		// Case 1: ${VAR:-default}
		if strings.Contains(match, ":-") {
			parts := strings.Split(match[2:len(match)-1], ":-")
			value := os.Getenv(parts[0])
			if value != "" {
				return value
			}
			return parts[1]
		}

		// Case 2: ${VAR} or $VAR
		varName := match
		if strings.HasPrefix(match, "${") {
			varName = match[2 : len(match)-1]
		} else {
			varName = match[1:]
		}

		if value := os.Getenv(varName); value != "" {
			return value
		}

		// Return original if not found
		return match
	})
}

// ProcessConfig reads the config file, processes environment variables, and returns a path to the processed config
func ProcessConfig(cfgFile string) (string, error) {
	content, err := os.ReadFile(cfgFile)
	if err != nil {
		return "", err
	}

	// Check if envfile is specified
	var cfg map[string]interface{}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return "", err
	}

	envFile, ok := cfg["envfile"].(string)
	if !ok || envFile == "" {
		return "", nil
	}

	// Load env file if specified
	if !filepath.IsAbs(envFile) {
		envFile = filepath.Join(filepath.Dir(cfgFile), envFile)
	}

	err = godotenv.Load(envFile)
	if err != nil {
		return "", err
	}

	// Perform environment variable substitution
	expandedContent := expandEnvVars(string(content))

	// Create temporary file with processed content
	tmpFile, err := os.CreateTemp("", "rr-processed-*.yaml")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tmpFile.Close()
	}()

	if err = os.WriteFile(tmpFile.Name(), []byte(expandedContent), 0644); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}
