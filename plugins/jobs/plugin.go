package jobs

import (
	"context"
	"fmt"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/server"
)

const (
	// RrJobs env variable
	RrJobs     string = "rr_jobs"
	PluginName string = "jobs"
)

type Plugin struct {
	cfg *Config
	log logger.Logger

	workersPool pool.Pool

	consumers map[string]Consumer
	events    events.Handler
}

func testListener(data interface{}) {
	fmt.Println(data)
}

func (p *Plugin) Init(cfg config.Configurer, log logger.Logger, server server.Server) error {
	const op = errors.Op("jobs_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &p.cfg)
	if err != nil {
		return errors.E(op, err)
	}

	err = p.cfg.InitDefaults()
	if err != nil {
		return errors.E(op, err)
	}

	p.workersPool, err = server.NewWorkerPool(context.Background(), p.cfg.poolCfg, map[string]string{RrJobs: "true"}, testListener)
	if err != nil {
		return errors.E(op, err)
	}

	p.events = events.NewEventsHandler()
	p.events.AddListener(testListener)
	p.consumers = make(map[string]Consumer)
	p.log = log
	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	// initialize sub-plugins
	// provide a queue to them
	// start consume loop
	// start resp loop

	return errCh
}

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.CollectMQBrokers,
	}
}

func (p *Plugin) CollectMQBrokers(name endure.Named, c Consumer) {
	p.consumers[name.Name()] = c
}

func (p *Plugin) Available() {}

func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Push(j *structs.Job) (string, error) {
	pipe, pOpts, err := p.cfg.MatchPipeline(j)
	if err != nil {
		panic(err)
	}

	if pOpts != nil {
		j.Options.Merge(pOpts)
	}

	broker, ok := p.consumers[pipe.Broker()]
	if !ok {
		panic("broker not found")
	}

	id, err := broker.Push(pipe, j)
	if err != nil {
		panic(err)
	}

	// p.events.Push()

	return id, nil
}

func (p *Plugin) RPC() interface{} {
	return &rpc{log: p.log}
}
