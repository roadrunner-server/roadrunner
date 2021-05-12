package broadcast

import (
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	PluginName string = "broadcast"
)

type Plugin struct {
	broker Subscriber
	driver Storage

	log logger.Logger
	cfg *Config
}

func (p *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("broadcast_plugin_init")

	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &p.cfg)
	if err != nil {
		return errors.E(op, errors.Disabled, err)
	}

	p.cfg.InitDefaults()

	p.log = log
	return nil
}

func (p *Plugin) Serve() chan error {
	const op = errors.Op("broadcast_plugin_serve")
	errCh := make(chan error)

	// if there are no brokers, return nil
	if p.broker == nil {
		errCh <- errors.E(op, errors.Str("no broker detected"))
		return errCh
	}

	// start the underlying broker
	go func() {
		// err := p.broker.Serve()
		// if err != nil {
		// 	errCh <- errors.E(op, err)
		// }
	}()

	return errCh
}

func (p *Plugin) Stop() error {
	return nil
}

// Available interface implementation for the plugin
func (p *Plugin) Available() {}

// Name is endure.Named interface implementation
func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.CollectBroker,
	}
}

func (p *Plugin) CollectBroker(name endure.Named, broker Subscriber) {
	p.broker = broker
}

func (p *Plugin) RPC() interface{} {
	// create an RPC service for the collects
	r := &rpc{
		log: p.log,
		svc: p,
	}
	return r
}

func (p *Plugin) Publish(msg []*Message) error {
	const op = errors.Op("broadcast_plugin_publish")
	return nil
}
