package broadcast

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	PluginName string = "broadcast"
	// driver is the mandatory field which should present in every storage
	driver string = "driver"

	redis  string = "redis"
	memory string = "memory"
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
	return make(chan error)
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
	go func() {
		p.Lock()
		defer p.Unlock()
		// check if any publisher registered
		if len(p.publishers) > 0 {
			for j := range p.publishers {
				err := p.publishers[j].Publish(m)
				if err != nil {
					p.log.Error("publishAsync", "error", err)
					// continue publish to other registered publishers
					continue
				}
			}
		} else {
			p.log.Warn("no publishers registered")
		}
	}()
}

func (p *Plugin) GetDriver(key string) (pubsub.SubReader, error) { //nolint:gocognit
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

		// config key for the particular sub-driver kv.memcached
		configKey := fmt.Sprintf("%s.%s", PluginName, key)

		switch val.(map[string]interface{})[driver] {
		case memory:
			if _, ok := p.constructors[memory]; !ok {
				return nil, errors.E(op, errors.Errorf("no memory drivers registered, registered: %s", p.publishers))
			}
			ps, err := p.constructors[memory].PSConstruct(configKey)
			if err != nil {
				return nil, errors.E(op, err)
			}

			// save the initialized publisher channel
			// for the in-memory, register new publishers
			p.publishers[uuid.NewString()] = ps

			return ps, nil
		case redis:
			if _, ok := p.constructors[redis]; !ok {
				return nil, errors.E(op, errors.Errorf("no redis drivers registered, registered: %s", p.publishers))
			}

			// first - try local configuration
			switch {
			case p.cfgPlugin.Has(configKey):
				ps, err := p.constructors[redis].PSConstruct(configKey)
				if err != nil {
					return nil, errors.E(op, err)
				}

				// if section already exists, return new connection
				if _, ok := p.publishers[configKey]; ok {
					return ps, nil
				}

				// if not - initialize a connection
				p.publishers[configKey] = ps
				return ps, nil

				// then try global if local does not exist
			case p.cfgPlugin.Has(redis):
				ps, err := p.constructors[redis].PSConstruct(configKey)
				if err != nil {
					return nil, errors.E(op, err)
				}

				// if section already exists, return new connection
				if _, ok := p.publishers[configKey]; ok {
					return ps, nil
				}

				// if not - initialize a connection
				p.publishers[configKey] = ps
				return ps, nil
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
