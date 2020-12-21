package config

import (
	"bytes"
	"errors"
	"strings"

	"github.com/spf13/viper"
)

type Viper struct {
	viper     *viper.Viper
	Path      string
	Prefix    string
	ReadInCfg []byte
}

// Inits config provider.
func (v *Viper) Init() error {
	v.viper = viper.New()

	// read in environment variables that match
	v.viper.AutomaticEnv()
	if v.Prefix == "" {
		return errors.New("prefix should be set")
	}

	v.viper.SetEnvPrefix(v.Prefix)
	if v.Path == "" {
		return errors.New("path should be set")
	}

	v.viper.SetConfigFile(v.Path)
	v.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if v.ReadInCfg != nil {
		return v.viper.ReadConfig(bytes.NewBuffer(v.ReadInCfg))
	}
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
	err := v.viper.UnmarshalKey(name, &out)
	if err != nil {
		return err
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
