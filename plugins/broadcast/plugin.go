package broadcast

import (
	"fmt"
	"sync"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	websocketsv1beta "github.com/spiral/roadrunner/v2/proto/websockets/v1beta"
	"google.golang.org/protobuf/proto"
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
	const op = errors.Op("broadcast_plugin_serve")
	errCh := make(chan error, 1)

	// iterate over config
	for k, v := range p.cfg.Data {
		if v == nil {
			continue
		}

		// check type of the v
		// should be a map[string]interface{}
		switch t := v.(type) {
		// correct type
		case map[string]interface{}:
			if _, ok := t[driver]; !ok {
				errCh <- errors.E(op, errors.Errorf("could not find mandatory driver field in the %s storage", k))
				return errCh
			}
		default:
			p.log.Warn("wrong type detected in the configuration, please, check yaml indentation")
			continue
		}

		// config key for the particular sub-driver kv.memcached
		configKey := fmt.Sprintf("%s.%s", PluginName, k)

		switch v.(map[string]interface{})[driver] {
		case memory:
			if _, ok := p.constructors[memory]; !ok {
				p.log.Warn("no memory drivers registered", "registered", p.publishers)
				continue
			}
			ps, err := p.constructors[memory].PSConstruct(configKey)
			if err != nil {
				errCh <- errors.E(op, err)
				return errCh
			}

			// save the pubsub
			p.publishers[k] = ps
		case redis:
			if _, ok := p.constructors[redis]; !ok {
				p.log.Warn("no redis drivers registered", "registered", p.publishers)
				continue
			}

			// first - try local configuration
			switch {
			case p.cfgPlugin.Has(configKey):
				ps, err := p.constructors[redis].PSConstruct(configKey)
				if err != nil {
					errCh <- errors.E(op, err)
					return errCh
				}

				// save the pubsub
				p.publishers[k] = ps
			case p.cfgPlugin.Has(redis):
				ps, err := p.constructors[redis].PSConstruct(configKey)
				if err != nil {
					errCh <- errors.E(op, err)
					return errCh
				}

				// save the pubsub
				p.publishers[k] = ps
				continue
			}
		}
	}

	return errCh
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
func (p *Plugin) CollectPublishers(name endure.Named, subscriber pubsub.Constructor) {
	// key redis, value - interface
	p.constructors[name.Name()] = subscriber
}

// Publish is an entry point to the websocket PUBSUB
func (p *Plugin) Publish(m []byte) error {
	p.Lock()
	defer p.Unlock()

	const op = errors.Op("broadcast_plugin_publish")

	msg := &websocketsv1beta.Message{}
	err := proto.Unmarshal(m, msg)
	if err != nil {
		return errors.E(op, err)
	}

	// Get payload
	for i := 0; i < len(msg.GetTopics()); i++ {
		if len(p.publishers) > 0 {
			for j := range p.publishers {
				err = p.publishers[j].Publish(m)
				if err != nil {
					return errors.E(op, err)
				}
			}

			return nil
		}

		p.log.Warn("no publishers registered")
	}

	return nil
}

func (p *Plugin) PublishAsync(m []byte) {
	go func() {
		p.Lock()
		defer p.Unlock()
		msg := &websocketsv1beta.Message{}
		err := proto.Unmarshal(m, msg)
		if err != nil {
			p.log.Error("message unmarshal")
		}

		// Get payload
		for i := 0; i < len(msg.GetTopics()); i++ {
			if len(p.publishers) > 0 {
				for j := range p.publishers {
					p.publishers[j].PublishAsync(m)
				}
				return
			}
			p.log.Warn("no publishers registered")
		}
	}()
}

func (p *Plugin) GetDriver(key string) (pubsub.SubReader, error) {
	const op = errors.Op("broadcast_plugin_get_driver")
	// key - driver, default for example
	// we should find `default` in the collected pubsubs constructors
	if pub, ok := p.publishers[key]; ok {
		return pub, nil
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
