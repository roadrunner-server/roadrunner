package memcached

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/kv"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const PluginName = "memcached"

var EmptyItem = kv.Item{}

type Plugin struct {
	// config plugin
	cfgPlugin config.Configurer
	// logger
	log logger.Logger
}

func (s *Plugin) Init(log logger.Logger, cfg config.Configurer) error {
	if !cfg.Has(kv.PluginName) {
		return errors.E(errors.Disabled)
	}

	s.cfgPlugin = cfg
	s.log = log
	return nil
}

// Name returns plugin user-friendly name
func (s *Plugin) Name() string {
	return PluginName
}

// Available interface implementation
func (s *Plugin) Available() {}

func (s *Plugin) Provide(key string) (kv.Storage, error) {
	const op = errors.Op("boltdb_plugin_provide")
	st, err := NewMemcachedDriver(s.log, key, s.cfgPlugin)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return st, nil
}
