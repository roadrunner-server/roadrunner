package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type ViperProvider struct {
	viper  *viper.Viper
	Path   string
	Prefix string
}

//////// ENDURE //////////
func (v *ViperProvider) Init() error {
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
	return v.viper.ReadInConfig()
}

///////////// VIPER ///////////////

// Overwrite overwrites existing config with provided values
func (v *ViperProvider) Overwrite(values map[string]string) error {
	if len(values) != 0 {
		for _, flag := range values {
			key, value, err := parseFlag(flag)
			if err != nil {
				return err
			}
			v.viper.Set(key, value)
		}
	}

	return nil
}

//
func (v *ViperProvider) UnmarshalKey(name string, out interface{}) error {
	err := v.viper.UnmarshalKey(name, &out)
	if err != nil {
		return err
	}
	return nil
}

// Get raw config in a form of config section.
func (v *ViperProvider) Get(name string) interface{} {
	return v.viper.Get(name)
}

/////////// PRIVATE //////////////

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
