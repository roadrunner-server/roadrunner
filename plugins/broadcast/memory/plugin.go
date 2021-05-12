package memory

import (
	"fmt"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/broadcast"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	PluginName  string = "broadcast"
	SectionName string = "memory"
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

	if !cfg.Has(fmt.Sprintf("%s.%s", PluginName, SectionName)) {
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
	// broadcast.memory
	return fmt.Sprintf("%s.%s", PluginName, SectionName)
}

func (p *Plugin) Publish(msg []*broadcast.Message) error {
	return nil
}
