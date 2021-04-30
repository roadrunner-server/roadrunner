package informer

import (
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/process"
)

const PluginName = "informer"

type Plugin struct {
	registry  map[string]Informer
	available map[string]Availabler
}

func (p *Plugin) Init() error {
	p.available = make(map[string]Availabler)
	p.registry = make(map[string]Informer)
	return nil
}

// Workers provides BaseProcess slice with workers for the requested plugin
func (p *Plugin) Workers(name string) ([]process.State, error) {
	const op = errors.Op("informer_plugin_workers")
	svc, ok := p.registry[name]
	if !ok {
		return nil, errors.E(op, errors.Errorf("no such service: %s", name))
	}

	return svc.Workers(), nil
}

// Collects declares services to be collected.
func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.CollectPlugins,
		p.CollectWorkers,
	}
}

// CollectPlugins collects all RR plugins
func (p *Plugin) CollectPlugins(name endure.Named, l Availabler) {
	p.available[name.Name()] = l
}

// CollectWorkers obtains plugins with workers inside.
func (p *Plugin) CollectWorkers(name endure.Named, r Informer) {
	p.registry[name.Name()] = r
}

// Name of the service.
func (p *Plugin) Name() string {
	return PluginName
}

// RPC returns associated rpc service.
func (p *Plugin) RPC() interface{} {
	return &rpc{srv: p}
}
