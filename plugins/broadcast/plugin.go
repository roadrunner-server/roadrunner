package broadcast

import (
	"sync"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	websocketsv1 "github.com/spiral/roadrunner/v2/pkg/proto/websockets/v1beta"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"google.golang.org/protobuf/proto"
)

const PluginName string = "broadcast"

type Plugin struct {
	sync.RWMutex
	log logger.Logger
	// publishers implement Publisher interface
	// and able to receive a payload
	publishers map[string]pubsub.Publisher
}

func (p *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("broadcast_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	p.publishers = make(map[string]pubsub.Publisher)
	p.log = log
	return nil
}

func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.CollectPublishers,
	}
}

// CollectPublishers collect all plugins who implement pubsub.Publisher interface
func (p *Plugin) CollectPublishers(name endure.Named, subscriber pubsub.Publisher) {
	p.publishers[name.Name()] = subscriber
}

// Publish is an entry point to the websocket PUBSUB
func (p *Plugin) Publish(m []byte) error {
	p.Lock()
	defer p.Unlock()

	const op = errors.Op("broadcast_plugin_publish")

	msg := &websocketsv1.Message{}
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
		msg := &websocketsv1.Message{}
		err := proto.Unmarshal(m, msg)
		if err != nil {
			p.log.Error("message unmarshal")
		}

		// Get payload
		for i := 0; i < len(msg.GetTopics()); i++ {
			if br, ok := p.publishers[msg.GetBroker()]; ok {
				err := br.Publish(m)
				if err != nil {
					p.log.Error("publish async error", "error", err)
				}
			} else {
				p.log.Warn("no such broker", "available", p.publishers, "requested", msg.GetBroker())
			}
		}
	}()
}

func (p *Plugin) GetDriver(key string) pubsub.SubReader {
	println(key)
	return nil
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
