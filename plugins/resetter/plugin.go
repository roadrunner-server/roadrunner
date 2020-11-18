package resetter

import (
	"github.com/spiral/endure"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/interfaces/log"
	"github.com/spiral/roadrunner/v2/interfaces/resetter"
)

const PluginName = "resetter"

type Plugin struct {
	registry map[string]resetter.Resetter
	log      log.Logger
}

func (p *Plugin) Init(log log.Logger) error {
	p.registry = make(map[string]resetter.Resetter)
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
func (p *Plugin) RegisterTarget(name endure.Named, r resetter.Resetter) error {
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

// RPCService returns associated rpc service.
func (p *Plugin) RPC() interface{} {
	return &rpc{srv: p, log: p.log}
}
