package memory

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/kv"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

// PluginName is user friendly name for the plugin
const PluginName = "memory"

type Plugin struct {
	// heap is user map for the key-value pairs
	stop chan struct{}

	log       logger.Logger
	cfgPlugin config.Configurer
	drivers   uint
}

func (s *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("in_memory_plugin_init")
	if !cfg.Has(kv.PluginName) {
		return errors.E(op, errors.Disabled)
	}

	s.log = log
	s.cfgPlugin = cfg
	s.stop = make(chan struct{}, 1)
	return nil
}

func (s *Plugin) Serve() chan error {
	return make(chan error, 1)
}

func (s *Plugin) Stop() error {
	if s.drivers > 0 {
		for i := uint(0); i < s.drivers; i++ {
			// send close signal to every driver
			s.stop <- struct{}{}
		}
	}
	return nil
}

func (s *Plugin) Provide(key string) (kv.Storage, error) {
	const op = errors.Op("inmemory_plugin_provide")
	st, err := NewInMemoryDriver(s.log, key, s.cfgPlugin, s.stop)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// save driver number to release resources after Stop
	s.drivers++

	return st, nil
}

// Name returns plugin user-friendly name
func (s *Plugin) Name() string {
	return PluginName
}

// Available interface implementation
func (s *Plugin) Available() {
}
