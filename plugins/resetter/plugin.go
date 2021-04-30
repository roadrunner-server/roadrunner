package resetter

import (
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const PluginName = "resetter"

type Plugin struct {
	registry map[string]Resettable
	log      logger.Logger
}

func (p *Plugin) ResetAll() error {
	const op = errors.Op("resetter_plugin_reset_all")
	for name := range p.registry {
		err := p.registry[name].Reset()
		if err != nil {
			return errors.E(op, err)
		}
	}
	return nil
}

func (p *Plugin) ResetByName(plugin string) error {
	const op = errors.Op("resetter_plugin_reset_by_name")
	if plugin, ok := p.registry[plugin]; ok {
		return plugin.Reset()
	}
	return errors.E(op, errors.Errorf("can't find plugin: %s", plugin))
}

func (p *Plugin) GetAll() []string {
	all := make([]string, 0, len(p.registry))
	for name := range p.registry {
		all = append(all, name)
	}
	return all
}

func (p *Plugin) Init(log logger.Logger) error {
	p.registry = make(map[string]Resettable)
	p.log = log
	return nil
}

// Reset named service.
func (p *Plugin) Reset(name string) error {
	svc, ok := p.registry[name]
	if !ok {
		return errors.E("no such service", errors.Str(name))
	}

	return svc.Reset()
}

// RegisterTarget resettable service.
func (p *Plugin) RegisterTarget(name endure.Named, r Resettable) error {
	p.registry[name.Name()] = r
	return nil
}

// Collects declares services to be collected.
func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.RegisterTarget,
	}
}

// Name of the service.
func (p *Plugin) Name() string {
	return PluginName
}

// Available interface implementation
func (p *Plugin) Available() {
}

// RPC returns associated rpc service.
func (p *Plugin) RPC() interface{} {
	return &rpc{srv: p, log: p.log}
}
