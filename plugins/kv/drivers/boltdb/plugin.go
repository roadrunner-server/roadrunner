package boltdb

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/kv"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const PluginName = "boltdb"

// Plugin BoltDB K/V storage.
type Plugin struct {
	cfgPlugin config.Configurer
	// logger
	log logger.Logger
	// stop is used to stop keys GC and close boltdb connection
	stop chan struct{}

	drivers uint
}

func (s *Plugin) Init(log logger.Logger, cfg config.Configurer) error {
	if !cfg.Has(kv.PluginName) {
		return errors.E(errors.Disabled)
	}

	s.stop = make(chan struct{})
	s.log = log
	s.cfgPlugin = cfg
	return nil
}

// Serve is noop here
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
	const op = errors.Op("boltdb_plugin_provide")
	st, err := NewBoltDBDriver(s.log, key, s.cfgPlugin, s.stop)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// save driver number to release resources after Stop
	s.drivers++

	return st, nil
}

// Name returns plugin name
func (s *Plugin) Name() string {
	return PluginName
}

// Available interface implementation
func (s *Plugin) Available() {}
