package informer

import (
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/process"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const PluginName = "informer"

type Plugin struct {
	registry map[string]Informer
	log      logger.Logger
}

func (p *Plugin) Init(log logger.Logger) error {
	p.registry = make(map[string]Informer)
	p.log = log
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

// CollectTarget resettable service.
func (p *Plugin) CollectTarget(name endure.Named, r Informer) error {
	p.registry[name.Name()] = r
	return nil
}

// Collects declares services to be collected.
func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.CollectTarget,
	}
}

// Name of the service.
func (p *Plugin) Name() string {
	return PluginName
}

// RPC returns associated rpc service.
func (p *Plugin) RPC() interface{} {
	return &rpc{srv: p, log: p.log}
}
