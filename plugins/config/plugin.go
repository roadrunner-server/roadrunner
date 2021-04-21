package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/spiral/errors"
)

type Viper struct {
	viper     *viper.Viper
	Path      string
	Prefix    string
	Type      string
	ReadInCfg []byte
	// user defined Flags in the form of <option>.<key> = <value>
	// which overwrites initial config key
	Flags []string

	CommonConfig *General
}

// Init config provider.
func (v *Viper) Init() error {
	const op = errors.Op("config_plugin_init")
	v.viper = viper.New()
	// If user provided []byte data with config, read it and ignore Path and Prefix
	if v.ReadInCfg != nil && v.Type != "" {
		v.viper.SetConfigType("yaml")
		return v.viper.ReadConfig(bytes.NewBuffer(v.ReadInCfg))
	}

	// read in environment variables that match
	v.viper.AutomaticEnv()
	if v.Prefix == "" {
		return errors.E(op, errors.Str("prefix should be set"))
	}

	v.viper.SetEnvPrefix(v.Prefix)
	if v.Path == "" {
		return errors.E(op, errors.Str("path should be set"))
	}

	v.viper.SetConfigFile(v.Path)
	v.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := v.viper.ReadInConfig()
	if err != nil {
		return errors.E(op, err)
	}

	// automatically inject ENV variables using ${ENV} pattern
	for _, key := range v.viper.AllKeys() {
		val := v.viper.Get(key)
		v.viper.Set(key, parseEnv(val))
	}

	// override config Flags
	if len(v.Flags) > 0 {
		for _, f := range v.Flags {
			key, val, err := parseFlag(f)
			if err != nil {
				return errors.E(op, err)
			}

			v.viper.Set(key, val)
		}
	}

	return nil
}

// Overwrite overwrites existing config with provided values
func (v *Viper) Overwrite(values map[string]interface{}) error {
	if len(values) != 0 {
		for key, value := range values {
			v.viper.Set(key, value)
		}
	}

	return nil
}

// UnmarshalKey reads configuration section into configuration object.
func (v *Viper) UnmarshalKey(name string, out interface{}) error {
	const op = errors.Op("config_plugin_unmarshal_key")
	err := v.viper.UnmarshalKey(name, &out)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (v *Viper) Unmarshal(out interface{}) error {
	const op = errors.Op("config_plugin_unmarshal")
	err := v.viper.Unmarshal(&out)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// Get raw config in a form of config section.
func (v *Viper) Get(name string) interface{} {
	return v.viper.Get(name)
}

// Has checks if config section exists.
func (v *Viper) Has(name string) bool {
	return v.viper.IsSet(name)
}

// Returns common config parameters
func (v *Viper) GetCommonConfig() *General {
	return v.CommonConfig
}

func parseFlag(flag string) (string, string, error) {
	const op = errors.Op("parse_flag")
	if !strings.Contains(flag, "=") {
		return "", "", errors.E(op, errors.Errorf("invalid flag `%s`", flag))
	}

	parts := strings.SplitN(strings.TrimLeft(flag, " \"'`"), "=", 2)

	return strings.Trim(parts[0], " \n\t"), parseValue(strings.Trim(parts[1], " \n\t")), nil
}

func parseValue(value string) string {
	escape := []rune(value)[0]

	if escape == '"' || escape == '\'' || escape == '`' {
		value = strings.Trim(value, string(escape))
		value = strings.ReplaceAll(value, fmt.Sprintf("\\%s", string(escape)), string(escape))
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
