package memory

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/kv"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const PluginName string = "memory"

type Plugin struct {
	// heap is user map for the key-value pairs
	stop chan struct{}

	log       logger.Logger
	cfgPlugin config.Configurer
	drivers   uint
}

func (p *Plugin) Init(log logger.Logger, cfg config.Configurer) error {
	p.log = log
	p.log = log
	p.cfgPlugin = cfg
	p.stop = make(chan struct{}, 1)
	return nil
}

func (p *Plugin) Serve() chan error {
	return make(chan error, 1)
}

func (p *Plugin) Stop() error {
	if p.drivers > 0 {
		for i := uint(0); i < p.drivers; i++ {
			// send close signal to every driver
			p.stop <- struct{}{}
		}
	}
	return nil
}

func (p *Plugin) PSConstruct(key string) (pubsub.PubSub, error) {
	return NewPubSubDriver(p.log, key)
}

func (p *Plugin) KVConstruct(key string) (kv.Storage, error) {
	const op = errors.Op("inmemory_plugin_provide")
	st, err := NewInMemoryDriver(p.log, key, p.cfgPlugin, p.stop)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// save driver number to release resources after Stop
	p.drivers++

	return st, nil
}

func (p *Plugin) Name() string {
	return PluginName
}
