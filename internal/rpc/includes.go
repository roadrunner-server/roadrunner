package rpc

import (
	"github.com/roadrunner-server/errors"
	"github.com/spf13/viper"
)

const (
	versionKey           string = "version"
	includeKey           string = "include"
	defaultConfigVersion string = "3"
	prevConfigVersion    string = "2.7"
)

func getConfiguration(path string) (map[string]any, string, error) {
	v := viper.New()
	v.SetConfigFile(path)
	err := v.ReadInConfig()
	if err != nil {
		return nil, "", err
	}

	// get configuration version
	ver := v.Get(versionKey)
	if ver == nil {
		return nil, "", errors.Str("rr configuration file should contain a version e.g: version: 2.7")
	}

	if _, ok := ver.(string); !ok {
		return nil, "", errors.Errorf("type of version should be string, actual: %T", ver)
	}

	// automatically inject ENV variables using ${ENV} pattern
	expandEnvViper(v)

	return v.AllSettings(), ver.(string), nil
}

func handleInclude(rootVersion string, v *viper.Viper) error {
	ifiles := v.GetStringSlice(includeKey)
	if ifiles == nil {
		return nil
	}

	for _, file := range ifiles {
		config, version, err := getConfiguration(file)
		if err != nil {
			return err
		}

		if version != rootVersion {
			return errors.Str("version in included file must be the same as in root")
		}

		// overriding configuration
		for key, val := range config {
			v.Set(key, val)
		}
	}

	return nil
}
