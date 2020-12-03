package informer

import (
	"github.com/spiral/endure"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2"
	"github.com/spiral/roadrunner/v2/interfaces/informer"
	"github.com/spiral/roadrunner/v2/interfaces/log"
)

const PluginName = "informer"

type Plugin struct {
	registry map[string]informer.Informer
	log      log.Logger
}

func (p *Plugin) Init(log log.Logger) error {
	p.registry = make(map[string]informer.Informer)
	p.log = log
	return nil
}

// Workers provides WorkerBase slice with workers for the requested plugin
func (p *Plugin) Workers(name string) ([]roadrunner.WorkerBase, error) {
	const op = errors.Op("get workers")
	svc, ok := p.registry[name]
	if !ok {
		return nil, errors.E(op, errors.Errorf("no such service: %s", name))
	}

	return svc.Workers(), nil
}

// CollectTarget resettable service.
func (p *Plugin) CollectTarget(name endure.Named, r informer.Informer) error {
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

// RPCService returns associated rpc service.
func (p *Plugin) RPC() interface{} {
	return &rpc{srv: p, log: p.log}
}
