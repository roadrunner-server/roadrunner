package jobs

import (
	"context"
	"fmt"
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

	pipelines string = "pipelines"
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

	// parent config for broken options. keys are pipelines names, values - pointers to the associated pipeline
	pipelines sync.Map

	// initial set of the pipelines to consume
	consume map[string]struct{}
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
	p.consume = make(map[string]struct{})

	// initial set of pipelines
	for i := range p.cfg.Pipelines {
		p.pipelines.Store(i, p.cfg.Pipelines[i])
	}

	if len(p.cfg.Consume) > 0 {
		for i := 0; i < len(p.cfg.Consume); i++ {
			p.consume[p.cfg.Consume[i]] = struct{}{}
		}
	}

	// initialize priority queue
	p.queue = priorityqueue.NewBinHeap()
	p.log = log

	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	const op = errors.Op("jobs_plugin_serve")

	// register initial pipelines
	p.pipelines.Range(func(key, value interface{}) bool {
		// pipeline name (ie test-local, sqs-aws, etc)
		name := key.(string)

		// skip pipelines which are not initialized to consume
		if _, ok := p.consume[name]; !ok {
			return true
		}

		// pipeline associated with the name
		pipe := value.(*pipeline.Pipeline)
		// driver for the pipeline (ie amqp, ephemeral, etc)
		dr := pipe.Driver()

		// jobConstructors contains constructors for the drivers
		// we need here to initialize these drivers for the pipelines
		if c, ok := p.jobConstructors[dr]; ok {
			// config key for the particular sub-driver jobs.pipelines.test-local
			configKey := fmt.Sprintf("%s.%s.%s", PluginName, pipelines, name)

			// init the driver
			initializedDriver, err := c.JobsConstruct(configKey, p.queue)
			if err != nil {
				errCh <- errors.E(op, err)
				return false
			}

			// add driver to the set of the consumers (name - pipeline name, value - associated driver)
			p.consumers[name] = initializedDriver

			// register pipeline for the initialized driver
			err = initializedDriver.Register(pipe)
			if err != nil {
				errCh <- errors.E(op, errors.Errorf("pipe register failed for the driver: %s with pipe name: %s", pipe.Driver(), pipe.Name()))
				return false
			}
		}

		return true
	})

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

					ctx, err := job.Context()
					if err != nil {
						job.Nack()
						p.log.Error("job marshal context", "error", err)
					}

					exec := payload.Payload{
						Context: ctx,
						Body:    job.Body(),
					}

					_, err = p.workersPool.Exec(exec)
					if err != nil {
						job.Nack()
						p.log.Error("job execute", "error", err)
						continue
					}

					job.Ack()
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

func (p *Plugin) Push(j *structs.Job) error {
	const op = errors.Op("jobs_plugin_push")

	// get the pipeline for the job
	pipe, ok := p.pipelines.Load(j.Options.Pipeline)
	if !ok {
		return errors.E(op, errors.Errorf("no such pipeline, requested: %s", j.Options.Pipeline))
	}

	// type conversion
	ppl := pipe.(*pipeline.Pipeline)

	d, ok := p.consumers[ppl.Name()]
	if !ok {
		return errors.E(op, errors.Errorf("consumer not registered for the requested driver: %s", ppl.Driver()))
	}

	// if job has no priority, inherit it from the pipeline
	if j.Options.Priority == 0 {
		j.Options.Priority = ppl.Priority()
	}

	err := d.Push(j)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (p *Plugin) PushBatch(j []*structs.Job) error {
	const op = errors.Op("jobs_plugin_push")

	for i := 0; i < len(j); i++ {
		// get the pipeline for the job
		pipe, ok := p.pipelines.Load(j[i].Options.Pipeline)
		if !ok {
			return errors.E(op, errors.Errorf("no such pipeline, requested: %s", j[i].Options.Pipeline))
		}

		ppl := pipe.(*pipeline.Pipeline)

		d, ok := p.consumers[ppl.Name()]
		if !ok {
			return errors.E(op, errors.Errorf("consumer not registered for the requested driver: %s", ppl.Driver()))
		}

		// if job has no priority, inherit it from the pipeline
		if j[i].Options.Priority == 0 {
			j[i].Options.Priority = ppl.Priority()
		}

		err := d.Push(j[i])
		if err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}

func (p *Plugin) Pause(pipelines []string) {
	for i := 0; i < len(pipelines); i++ {
		pipe, ok := p.pipelines.Load(pipelines[i])
		if !ok {
			p.log.Error("no such pipeline", "requested", pipelines[i])
		}

		ppl := pipe.(*pipeline.Pipeline)

		d, ok := p.consumers[ppl.Name()]
		if !ok {
			p.log.Warn("driver for the pipeline not found", "pipeline", pipelines[i])
			return
		}

		// redirect call to the underlying driver
		d.Pause(ppl.Name())
	}
}

func (p *Plugin) Resume(pipelines []string) {
	for i := 0; i < len(pipelines); i++ {
		pipe, ok := p.pipelines.Load(pipelines[i])
		if !ok {
			p.log.Error("no such pipeline", "requested", pipelines[i])
		}

		ppl := pipe.(*pipeline.Pipeline)

		d, ok := p.consumers[ppl.Name()]
		if !ok {
			p.log.Warn("driver for the pipeline not found", "pipeline", pipelines[i])
			return
		}

		// redirect call to the underlying driver
		d.Resume(ppl.Name())
	}
}

func (p *Plugin) RPC() interface{} {
	return &rpc{
		log: p.log,
		p:   p,
	}
}
