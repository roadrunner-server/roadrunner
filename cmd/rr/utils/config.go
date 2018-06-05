package utils

import (
	"github.com/spf13/viper"
	"github.com/spiral/roadrunner/service"
)

// ViperWrapper provides interface bridge between Viper configs and service.Config.
type ViperWrapper struct {
	Viper *viper.Viper
}

// Get nested config section (sub-map), returns nil if section not found.
func (w *ViperWrapper) Get(key string) service.Config {
	sub := w.Viper.Sub(key)
	if sub == nil {
		return nil
	}

	return &ViperWrapper{sub}
}

// Unmarshal unmarshal config data into given struct.
func (w *ViperWrapper) Unmarshal(out interface{}) error {
	return w.Viper.Unmarshal(out)
}
