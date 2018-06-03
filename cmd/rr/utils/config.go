package utils

import (
	"github.com/spf13/viper"
	"github.com/spiral/roadrunner/service"
)

type ConfigWrapper struct {
	Viper *viper.Viper
}

func (w *ConfigWrapper) Get(key string) service.Config {
	sub := w.Viper.Sub(key)
	if sub == nil {
		return nil
	}

	return &ConfigWrapper{sub}
}

func (w *ConfigWrapper) Unmarshal(out interface{}) error {
	return w.Viper.Unmarshal(out)
}
