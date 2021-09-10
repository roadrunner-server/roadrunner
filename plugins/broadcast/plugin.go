package broadcast

import (
	"fmt"
	"sync"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/common/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	PluginName string = "broadcast"
	// driver is the mandatory field which should present in every storage
	driver string = "driver"

	// every driver should have config section for the local configuration
	conf string = "config"
)

type Plugin struct {
	sync.RWMutex

	cfg       *Config
	cfgPlugin config.Configurer
	log       logger.Logger
	// publishers implement Publisher interface
	// and able to receive a payload
	publishers   map[string]pubsub.PubSub
	constructors map[string]pubsub.Constructor
}

func (p *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("broadcast_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}
	p.cfg = &Config{}
	// unmarshal config section
	err := cfg.UnmarshalKey(PluginName, &p.cfg.Data)
	if err != nil {
		return errors.E(op, err)
	}

	p.publishers = make(map[string]pubsub.PubSub)
	p.constructors = make(map[string]pubsub.Constructor)

	p.log = log
	p.cfgPlugin = cfg
	return nil
}

func (p *Plugin) Serve() chan error {
	return make(chan error, 1)
}

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.CollectPublishers,
	}
}

// CollectPublishers collect all plugins who implement pubsub.Publisher interface
func (p *Plugin) CollectPublishers(name endure.Named, constructor pubsub.Constructor) {
	// key redis, value - interface
	p.constructors[name.Name()] = constructor
}

// Publish is an entry point to the websocket PUBSUB
func (p *Plugin) Publish(m *pubsub.Message) error {
	p.Lock()
	defer p.Unlock()

	const op = errors.Op("broadcast_plugin_publish")

	// check if any publisher registered
	if len(p.publishers) > 0 {
		for j := range p.publishers {
			err := p.publishers[j].Publish(m)
			if err != nil {
				return errors.E(op, err)
			}
		}
		return nil
	} else {
		p.log.Warn("no publishers registered")
	}

	return nil
}

func (p *Plugin) PublishAsync(m *pubsub.Message) {
	// TODO(rustatian) channel here?
	go func() {
		p.Lock()
		defer p.Unlock()
		// check if any publisher registered
		if len(p.publishers) > 0 {
			for j := range p.publishers {
				err := p.publishers[j].Publish(m)
				if err != nil {
					p.log.Error("publishAsync", "error", err)
					// continue publishing to the other registered publishers
					continue
				}
			}
		} else {
			p.log.Warn("no publishers registered")
		}
	}()
}

func (p *Plugin) GetDriver(key string) (pubsub.SubReader, error) {
	const op = errors.Op("broadcast_plugin_get_driver")

	// choose a driver
	if val, ok := p.cfg.Data[key]; ok {
		// check type of the v
		// should be a map[string]interface{}
		switch t := val.(type) {
		// correct type
		case map[string]interface{}:
			if _, ok := t[driver]; !ok {
				panic(errors.E(op, errors.Errorf("could not find mandatory driver field in the %s storage", val)))
			}
		default:
			return nil, errors.E(op, errors.Str("wrong type detected in the configuration, please, check yaml indentation"))
		}

		// config key for the particular sub-driver broadcast.memcached.config
		configKey := fmt.Sprintf("%s.%s.%s", PluginName, key, conf)

		drName := val.(map[string]interface{})[driver]

		// driver name should be a string
		if drStr, ok := drName.(string); ok {
			if _, ok := p.constructors[drStr]; !ok {
				return nil, errors.E(op, errors.Errorf("no drivers with the requested name registered, registered: %s, requested: %s", p.publishers, drStr))
			}

			switch {
			// try local config first
			case p.cfgPlugin.Has(configKey):
				// we found a local configuration
				ps, err := p.constructors[drStr].PSConstruct(configKey)
				if err != nil {
					return nil, errors.E(op, err)
				}

				// save the initialized publisher channel
				// for the in-memory, register new publishers
				p.publishers[configKey] = ps

				return ps, nil
			case p.cfgPlugin.Has(key):
				// try global driver section after local
				ps, err := p.constructors[drStr].PSConstruct(key)
				if err != nil {
					return nil, errors.E(op, err)
				}

				// save the initialized publisher channel
				// for the in-memory, register new publishers
				p.publishers[configKey] = ps

				return ps, nil
			default:
				p.log.Error("can't find local or global configuration, this section will be skipped", "local: ", configKey, "global: ", key)
			}
		}
	}
	return nil, errors.E(op, errors.Str("could not find driver by provided key"))
}

func (p *Plugin) RPC() interface{} {
	return &rpc{
		plugin: p,
		log:    p.log,
	}
}

func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Available() {}
