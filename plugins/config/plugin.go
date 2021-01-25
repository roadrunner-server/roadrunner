package config

import (
	"bytes"
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
}

// Inits config provider.
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

	return v.viper.ReadInConfig()
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
