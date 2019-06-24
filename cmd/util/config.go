package util

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/spiral/roadrunner/service"
	"os"
	"path/filepath"
	"strings"
)

// configWrapper provides interface bridge between v configs and service.Config.
type configWrapper struct {
	v *viper.Viper
}

// Get nested config section (sub-map), returns nil if section not found.
func (w *configWrapper) Get(key string) service.Config {
	sub := w.v.Sub(key)
	if sub == nil {
		return nil
	}

	return &configWrapper{sub}
}

// Unmarshal unmarshal config data into given struct.
func (w *configWrapper) Unmarshal(out interface{}) error {
	return w.v.Unmarshal(out)
}

// LoadConfig config and merge it's values with set of flags.
func LoadConfig(cfgFile string, path []string, name string, flags []string) (*configWrapper, error) {
	cfg := viper.New()

	if cfgFile != "" {
		if absPath, err := filepath.Abs(cfgFile); err == nil {
			cfgFile = absPath

			// force working absPath related to config file
			if err := os.Chdir(filepath.Dir(absPath)); err != nil {
				return nil, err
			}
		}

		// Use cfg file from the flag.
		cfg.SetConfigFile(cfgFile)

		if dir, err := filepath.Abs(cfgFile); err == nil {
			// force working absPath related to config file
			if err := os.Chdir(filepath.Dir(dir)); err != nil {
				return nil, err
			}
		}
	} else {
		// automatic location
		for _, p := range path {
			cfg.AddConfigPath(p)
		}

		cfg.SetConfigName(name)
	}

	// read in environment variables that match
	cfg.AutomaticEnv()
	cfg.SetEnvPrefix("rr")
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// If a cfg file is found, read it in.
	if err := cfg.ReadInConfig(); err != nil {
		if len(flags) == 0 {
			return nil, err
		}
	}

	// merge included configs
	if include, ok := cfg.Get("include").([]interface{}); ok {

		for _, file := range include {
			filename, ok := file.(string)
			if !ok {
				continue
			}

			partial := viper.New()
			partial.AutomaticEnv()
			partial.SetEnvPrefix("rr")
			partial.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
			partial.SetConfigFile(filename)

			if err := partial.ReadInConfig(); err != nil {
				return nil, err
			}

			// merging
			if err := cfg.MergeConfigMap(partial.AllSettings()); err != nil {
				return nil, err
			}
		}
	}

	// automatically inject ENV variables using ${ENV} pattern
	for _, key := range cfg.AllKeys() {
		val := cfg.Get(key)
		cfg.Set(key, parseEnv(val))
	}

	// merge with console flags
	if len(flags) != 0 {
		for _, f := range flags {
			k, v, err := parseFlag(f)
			if err != nil {
				return nil, err
			}

			cfg.Set(k, v)
		}
	}

	merged := viper.New()

	// we have to copy all the merged values into new config in order normalize it (viper bug?)
	if err := merged.MergeConfigMap(cfg.AllSettings()); err != nil {
		return nil, err
	}

	return &configWrapper{merged}, nil
}

func parseFlag(flag string) (string, string, error) {
	if !strings.Contains(flag, "=") {
		return "", "", fmt.Errorf("invalid flag `%s`", flag)
	}

	parts := strings.SplitN(strings.TrimLeft(flag, " \"'`"), "=", 2)

	return strings.Trim(parts[0], " \n\t"), parseValue(strings.Trim(parts[1], " \n\t")), nil
}

func parseValue(value string) string {
	escape := []rune(value)[0]

	if escape == '"' || escape == '\'' || escape == '`' {
		value = strings.Trim(value, string(escape))
		value = strings.Replace(value, fmt.Sprintf("\\%s", string(escape)), string(escape), -1)
	}

	return value
}

func parseEnv(value interface{}) interface{} {
	str, ok := value.(string)
	if !ok || len(str) <= 3 {
		return value
	}

	if str[0:2] == "${" && str[len(str)-1:] == "}" {
		if v, ok := os.LookupEnv(str[2 : len(str)-1]); ok {
			return v
		}
	}

	return str
}
