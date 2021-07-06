package jobs

import (
	"context"
	"sync"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/common/jobs"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/payload"
	"github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/pkg/priorityqueue"
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
	cfg *Config `mapstructure:"jobs"`
	log logger.Logger

	sync.RWMutex

	workersPool pool.Pool
	server      server.Server

	jobConstructors map[string]jobs.Constructor
	consumers       map[string]jobs.Consumer

	events events.Handler

	// priority queue implementation
	queue priorityqueue.Queue

	// parent config for broken options.
	pipelines pipeline.Pipelines
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
	p.jobConstructors = make(map[string]jobs.Constructor)
	p.consumers = make(map[string]jobs.Consumer)

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
	const op = errors.Op("jobs_plugin_serve")

	for name := range p.jobConstructors {
		jb, err := p.jobConstructors[name].JobsConstruct("", p.queue)
		if err != nil {
			errCh <- err
			return errCh
		}

		p.consumers[name] = jb
	}

	// register initial pipelines
	for i := 0; i < len(p.pipelines); i++ {
		pipe := p.pipelines[i]

		if jb, ok := p.consumers[pipe.Driver()]; ok {
			err := jb.Register(pipe.Name())
			if err != nil {
				errCh <- errors.E(op, errors.Errorf("pipe register failed for the driver: %s with pipe name: %s", pipe.Driver(), pipe.Name()))
				return errCh
			}
		}
	}

	var err error
	p.workersPool, err = p.server.NewWorkerPool(context.Background(), p.cfg.Pool, map[string]string{RrJobs: "true"})
	if err != nil {
		errCh <- err
		return errCh
	}

	// start listening
	go func() {
		for i := uint8(0); i < p.cfg.NumPollers; i++ {
			go func() {
				for {
					// get data JOB from the queue
					job := p.queue.GetMax()

					if job == nil {
						continue
					}

					exec := payload.Payload{
						Context: job.Context(),
						Body:    job.Body(),
					}

					_, err := p.workersPool.Exec(exec)
					if err != nil {
						panic(err)
					}
				}
			}()
		}
	}()

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

func (p *Plugin) CollectMQBrokers(name endure.Named, c jobs.Constructor) {
	p.jobConstructors[name.Name()] = c
}

func (p *Plugin) Available() {}

func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Push(j *structs.Job) (*string, error) {
	const op = errors.Op("jobs_plugin_push")
	pipe := p.pipelines.Get(j.Options.Pipeline)

	broker, ok := p.consumers[pipe.Driver()]
	if !ok {
		return nil, errors.E(op, errors.Errorf("consumer not registered for the requested driver: %s", pipe.Driver()))
	}

	id, err := broker.Push(j)
	if err != nil {
		panic(err)
	}

	return id, nil
}

func (p *Plugin) RPC() interface{} {
	return &rpc{
		log: p.log,
		p:   p,
	}
}
