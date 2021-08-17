package informer

import (
	"context"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner/v2/pkg/state/job"
	"github.com/spiral/roadrunner/v2/pkg/state/process"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const PluginName = "informer"

type Plugin struct {
	log logger.Logger

	withJobs    map[string]JobsStat
	withWorkers map[string]Informer
	available   map[string]Availabler
}

func (p *Plugin) Init(log logger.Logger) error {
	p.available = make(map[string]Availabler)
	p.withWorkers = make(map[string]Informer)
	p.withJobs = make(map[string]JobsStat)

	p.log = log
	return nil
}

// Workers provides BaseProcess slice with workers for the requested plugin
func (p *Plugin) Workers(name string) []*process.State {
	svc, ok := p.withWorkers[name]
	if !ok {
		return nil
	}

	return svc.Workers()
}

// Jobs provides information about jobs for the registered plugin using jobs
func (p *Plugin) Jobs(name string) []*job.State {
	svc, ok := p.withJobs[name]
	if !ok {
		return nil
	}

	st, err := svc.JobsState(context.Background())
	if err != nil {
		p.log.Info("jobs stat", "error", err)
		// skip errors here
		return nil
	}

	return st
}

// Collects declares services to be collected.
func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.CollectPlugins,
		p.CollectWorkers,
		p.CollectJobs,
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

func (p *Plugin) CollectJobs(name endure.Named, j JobsStat) {
	p.withJobs[name.Name()] = j
}

// Name of the service.
func (p *Plugin) Name() string {
	return PluginName
}

// RPC returns associated rpc service.
func (p *Plugin) RPC() interface{} {
	return &rpc{srv: p}
}
