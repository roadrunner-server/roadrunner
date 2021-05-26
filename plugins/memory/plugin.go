package memory

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	PluginName string = "memory"
)

type Plugin struct {
	log logger.Logger
	cfg *Config
}

func (p *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("memory_plugin_init")

	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	p.log = log
	return nil
}

func (p *Plugin) Serve() chan error {
	const op = errors.Op("memory_plugin_serve")
	errCh := make(chan error)

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

func (p *Plugin) Publish(messages []pubsub.Message) error {
	panic("implement me")
}

func (p *Plugin) PublishAsync(messages []pubsub.Message) {
	panic("implement me")
}

func (p *Plugin) Subscribe(topics ...string) error {
	panic("implement me")
}

func (p *Plugin) Unsubscribe(topics ...string) error {
	panic("implement me")
}

func (p *Plugin) Next() (pubsub.Message, error) {
	panic("implement me")
}
