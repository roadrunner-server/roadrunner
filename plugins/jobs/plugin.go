package jobs

import (
	"context"
	"fmt"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/pool"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
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
	server      server.Server

	brokers   map[string]Broker
	consumers map[string]Consumer

	events events.Handler

	// priority queue implementation
	queue priorityqueue.Queue

	// parent config for broken options.
	pipelines pipeline.Pipelines
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

	p.cfg.InitDefaults()

	p.server = server
	p.events = events.NewEventsHandler()
	p.events.AddListener(testListener)
	p.brokers = make(map[string]Broker)
	p.consumers = make(map[string]Consumer)

	// initial set of pipelines
	p.pipelines, err = pipeline.InitPipelines(p.cfg.Pipelines)
	if err != nil {
		return errors.E(op, err)
	}

	// initialize priority queue
	p.queue = priorityqueue.NewBinHeap()
	p.log = log

	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	for name := range p.brokers {
		jb, err := p.brokers[name].InitJobBroker(p.queue)
		if err != nil {
			errCh <- err
			return errCh
		}

		p.consumers[name] = jb
	}

	var err error
	p.workersPool, err = p.server.NewWorkerPool(context.Background(), p.cfg.poolCfg, map[string]string{RrJobs: "true"}, testListener)
	if err != nil {
		errCh <- err
		return errCh
	}

	// initialize sub-plugins
	// provide a queue to them
	// start consume loop
	// start resp loop

	/*
		go func() {
			for {
				// get data JOB from the queue
				job := p.queue.Pop()

				// request
				_ = job
				p.workersPool.Exec(nil)
			}
		}()

	*/
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

func (p *Plugin) CollectMQBrokers(name endure.Named, c Broker) {
	p.brokers[name.Name()] = c
}

func (p *Plugin) Available() {}

func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Push(j *structs.Job) (string, error) {
	pipe := p.pipelines.Get(j.Options.Pipeline)

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
	return &rpc{
		log: p.log,
		p:   p,
	}
}
