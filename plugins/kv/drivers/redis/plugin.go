package redis

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/kv"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const PluginName = "redis"

// Plugin BoltDB K/V storage.
type Plugin struct {
	cfgPlugin config.Configurer
	// logger
	log logger.Logger
}

func (s *Plugin) Init(log logger.Logger, cfg config.Configurer) error {
	if !cfg.Has(kv.PluginName) {
		return errors.E(errors.Disabled)
	}

	s.log = log
	s.cfgPlugin = cfg
	return nil
}

// Serve is noop here
func (s *Plugin) Serve() chan error {
	return make(chan error, 1)
}

func (s *Plugin) Stop() error {
	return nil
}

func (s *Plugin) Provide(key string) (kv.Storage, error) {
	const op = errors.Op("redis_plugin_provide")
	st, err := NewRedisDriver(s.log, key, s.cfgPlugin)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return st, nil
}

// Name returns plugin name
func (s *Plugin) Name() string {
	return PluginName
}
