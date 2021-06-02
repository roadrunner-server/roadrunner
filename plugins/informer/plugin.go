package informer

import (
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner/v2/pkg/process"
)

const PluginName = "informer"

type Plugin struct {
	withWorkers map[string]Informer
	available   map[string]Availabler
}

func (p *Plugin) Init() error {
	p.available = make(map[string]Availabler)
	p.withWorkers = make(map[string]Informer)
	return nil
}

// Workers provides BaseProcess slice with workers for the requested plugin
func (p *Plugin) Workers(name string) []process.State {
	svc, ok := p.withWorkers[name]
	if !ok {
		return nil
	}

	return svc.Workers()
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
	p.withWorkers[name.Name()] = r
}

// Name of the service.
func (p *Plugin) Name() string {
	return PluginName
}

// RPC returns associated rpc service.
func (p *Plugin) RPC() interface{} {
	return &rpc{srv: p}
}
