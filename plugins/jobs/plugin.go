package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/common/jobs"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/payload"
	"github.com/spiral/roadrunner/v2/pkg/pool"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	jobState "github.com/spiral/roadrunner/v2/pkg/state/job"
	"github.com/spiral/roadrunner/v2/pkg/state/process"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/job"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/server"
)

const (
	// RrMode env variable
	RrMode     string = "RR_MODE"
	RrModeJobs string = "jobs"

	PluginName string = "jobs"
	pipelines  string = "pipelines"
)

type Plugin struct {
	sync.RWMutex

	// Jobs plugin configuration
	cfg         *Config `structure:"jobs"`
	log         logger.Logger
	workersPool pool.Pool
	server      server.Server

	jobConstructors map[string]jobs.Constructor
	consumers       map[string]jobs.Consumer

	// events handler
	events events.Handler

	// priority queue implementation
	queue priorityqueue.Queue

	// parent config for broken options. keys are pipelines names, values - pointers to the associated pipeline
	pipelines sync.Map

	// initial set of the pipelines to consume
	consume map[string]struct{}

	// signal channel to stop the pollers
	stopCh chan struct{}

	// internal payloads pool
	pldPool       sync.Pool
	statsExporter *statsExporter
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
	p.events.AddListener(p.collectJobsEvents)

	p.jobConstructors = make(map[string]jobs.Constructor)
	p.consumers = make(map[string]jobs.Consumer)
	p.consume = make(map[string]struct{})
	p.stopCh = make(chan struct{}, 1)

	p.pldPool = sync.Pool{New: func() interface{} {
		// with nil fields
		return &payload.Payload{}
	}}

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
	p.queue = priorityqueue.NewBinHeap(p.cfg.PipelineSize)
	p.log = log

	// metrics
	p.statsExporter = newStatsExporter(p)
	p.events.AddListener(p.statsExporter.metricsCallback)

	return nil
}

func (p *Plugin) Serve() chan error { //nolint:gocognit
	errCh := make(chan error, 1)
	const op = errors.Op("jobs_plugin_serve")

	// register initial pipelines
	p.pipelines.Range(func(key, value interface{}) bool {
		t := time.Now()
		// pipeline name (ie test-local, sqs-aws, etc)
		name := key.(string)

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
			initializedDriver, err := c.JobsConstruct(configKey, p.events, p.queue)
			if err != nil {
				errCh <- errors.E(op, err)
				return false
			}

			// add driver to the set of the consumers (name - pipeline name, value - associated driver)
			p.consumers[name] = initializedDriver

			// register pipeline for the initialized driver
			err = initializedDriver.Register(context.Background(), pipe)
			if err != nil {
				errCh <- errors.E(op, errors.Errorf("pipe register failed for the driver: %s with pipe name: %s", pipe.Driver(), pipe.Name()))
				return false
			}

			// if pipeline initialized to be consumed, call Run on it
			if _, ok := p.consume[name]; ok {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
				defer cancel()
				err = initializedDriver.Run(ctx, pipe)
				if err != nil {
					errCh <- errors.E(op, err)
					return false
				}
				return true
			}

			return true
		}

		p.events.Push(events.JobEvent{
			Event:    events.EventDriverReady,
			Pipeline: pipe.Name(),
			Driver:   pipe.Driver(),
			Start:    t,
			Elapsed:  t.Sub(t),
		})

		return true
	})

	// do not continue processing, immediately stop if channel contains an error
	if len(errCh) > 0 {
		return errCh
	}

	var err error
	p.workersPool, err = p.server.NewWorkerPool(context.Background(), p.cfg.Pool, map[string]string{RrMode: RrModeJobs})
	if err != nil {
		errCh <- err
		return errCh
	}

	// start listening
	go func() {
		for i := uint8(0); i < p.cfg.NumPollers; i++ {
			go func() {
				for {
					select {
					case <-p.stopCh:
						p.log.Info("------> job poller stopped <------")
						return
					default:
						// get prioritized JOB from the queue
						jb := p.queue.ExtractMin()

						// parse the context
						// for each job, context contains:
						/*
							1. Job class
							2. Job ID provided from the outside
							3. Job Headers map[string][]string
							4. Timeout in seconds
							5. Pipeline name
						*/

						start := time.Now()
						p.events.Push(events.JobEvent{
							Event:   events.EventJobStart,
							ID:      jb.ID(),
							Start:   start,
							Elapsed: 0,
						})

						ctx, err := jb.Context()
						if err != nil {
							p.events.Push(events.JobEvent{
								Event:   events.EventJobError,
								Error:   err,
								ID:      jb.ID(),
								Start:   start,
								Elapsed: time.Since(start),
							})

							errNack := jb.Nack()
							if errNack != nil {
								p.log.Error("negatively acknowledge failed", "error", errNack)
							}
							p.log.Error("job marshal context", "error", err)
							continue
						}

						// get payload from the sync.Pool
						exec := p.getPayload(jb.Body(), ctx)

						// protect from the pool reset
						p.RLock()
						resp, err := p.workersPool.Exec(exec)
						p.RUnlock()
						if err != nil {
							p.events.Push(events.JobEvent{
								Event:   events.EventJobError,
								ID:      jb.ID(),
								Error:   err,
								Start:   start,
								Elapsed: time.Since(start),
							})
							// RR protocol level error, Nack the job
							errNack := jb.Nack()
							if errNack != nil {
								p.log.Error("negatively acknowledge failed", "error", errNack)
							}

							p.log.Error("job execute failed", "error", err)

							p.putPayload(exec)
							continue
						}

						// if response is nil or body is nil, just acknowledge the job
						if resp == nil || resp.Body == nil {
							p.putPayload(exec)
							err = jb.Ack()
							if err != nil {
								p.events.Push(events.JobEvent{
									Event:   events.EventJobError,
									ID:      jb.ID(),
									Error:   err,
									Start:   start,
									Elapsed: time.Since(start),
								})
								p.log.Error("acknowledge error, job might be missed", "error", err)
								continue
							}

							p.events.Push(events.JobEvent{
								Event:   events.EventJobOK,
								ID:      jb.ID(),
								Start:   start,
								Elapsed: time.Since(start),
							})

							continue
						}

						// handle the response protocol
						err = handleResponse(resp.Body, jb, p.log)
						if err != nil {
							p.events.Push(events.JobEvent{
								Event:   events.EventJobError,
								ID:      jb.ID(),
								Start:   start,
								Error:   err,
								Elapsed: time.Since(start),
							})
							p.putPayload(exec)
							errNack := jb.Nack()
							if errNack != nil {
								p.log.Error("negatively acknowledge failed, job might be lost", "root error", err, "error nack", errNack)
								continue
							}

							p.log.Error("job negatively acknowledged", "error", err)
							continue
						}

						p.events.Push(events.JobEvent{
							Event:   events.EventJobOK,
							ID:      jb.ID(),
							Start:   start,
							Elapsed: time.Since(start),
						})

						// return payload
						p.putPayload(exec)
					}
				}
			}()
		}
	}()

	return errCh
}

func (p *Plugin) Stop() error {
	for k, v := range p.consumers {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
		err := v.Stop(ctx)
		if err != nil {
			cancel()
			p.log.Error("stop job driver", "driver", k)
			continue
		}
		cancel()
	}

	// this function can block forever, but we don't care, because we might have a chance to exit from the pollers,
	// but if not, this is not a problem at all.
	// The main target is to stop the drivers
	go func() {
		for i := uint8(0); i < p.cfg.NumPollers; i++ {
			// stop jobs plugin pollers
			p.stopCh <- struct{}{}
		}
	}()

	// just wait pollers for 5 seconds before exit
	time.Sleep(time.Second * 5)

	p.Lock()
	p.workersPool.Destroy(context.Background())
	p.Unlock()

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

func (p *Plugin) Workers() []*process.State {
	p.RLock()
	wrk := p.workersPool.Workers()
	p.RUnlock()

	ps := make([]*process.State, len(wrk))

	for i := 0; i < len(wrk); i++ {
		st, err := process.WorkerProcessState(wrk[i])
		if err != nil {
			p.log.Error("jobs workers state", "error", err)
			return nil
		}

		ps[i] = st
	}

	return ps
}

func (p *Plugin) JobsState(ctx context.Context) ([]*jobState.State, error) {
	const op = errors.Op("jobs_plugin_drivers_state")
	jst := make([]*jobState.State, 0, len(p.consumers))
	for k := range p.consumers {
		d := p.consumers[k]
		newCtx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(p.cfg.Timeout))
		state, err := d.State(newCtx)
		if err != nil {
			cancel()
			return nil, errors.E(op, err)
		}

		jst = append(jst, state)
		cancel()
	}
	return jst, nil
}

func (p *Plugin) Available() {}

func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Reset() error {
	p.Lock()
	defer p.Unlock()

	const op = errors.Op("jobs_plugin_reset")
	p.log.Info("JOBS plugin received restart request. Restarting...")
	p.workersPool.Destroy(context.Background())
	p.workersPool = nil

	var err error
	p.workersPool, err = p.server.NewWorkerPool(context.Background(), p.cfg.Pool, map[string]string{RrMode: RrModeJobs}, p.collectJobsEvents, p.statsExporter.metricsCallback)
	if err != nil {
		return errors.E(op, err)
	}

	p.log.Info("JOBS workers pool successfully restarted")

	return nil
}

func (p *Plugin) Push(j *job.Job) error {
	const op = errors.Op("jobs_plugin_push")

	start := time.Now()
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
	defer cancel()

	err := d.Push(ctx, j)
	if err != nil {
		p.events.Push(events.JobEvent{
			Event:    events.EventPushError,
			ID:       j.Ident,
			Pipeline: ppl.Name(),
			Driver:   ppl.Driver(),
			Error:    err,
			Start:    start,
			Elapsed:  time.Since(start),
		})
		return errors.E(op, err)
	}

	p.events.Push(events.JobEvent{
		Event:    events.EventPushOK,
		ID:       j.Ident,
		Pipeline: ppl.Name(),
		Driver:   ppl.Driver(),
		Error:    err,
		Start:    start,
		Elapsed:  time.Since(start),
	})

	return nil
}

func (p *Plugin) PushBatch(j []*job.Job) error {
	const op = errors.Op("jobs_plugin_push")
	start := time.Now()

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

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
		err := d.Push(ctx, j[i])
		if err != nil {
			cancel()
			p.events.Push(events.JobEvent{
				Event:    events.EventPushError,
				ID:       j[i].Ident,
				Pipeline: ppl.Name(),
				Driver:   ppl.Driver(),
				Start:    start,
				Elapsed:  time.Since(start),
				Error:    err,
			})
			return errors.E(op, err)
		}

		cancel()
	}

	return nil
}

func (p *Plugin) Pause(pp string) {
	pipe, ok := p.pipelines.Load(pp)

	if !ok {
		p.log.Error("no such pipeline", "requested", pp)
	}

	ppl := pipe.(*pipeline.Pipeline)

	d, ok := p.consumers[ppl.Name()]
	if !ok {
		p.log.Warn("driver for the pipeline not found", "pipeline", pp)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
	defer cancel()
	// redirect call to the underlying driver
	d.Pause(ctx, ppl.Name())
}

func (p *Plugin) Resume(pp string) {
	pipe, ok := p.pipelines.Load(pp)
	if !ok {
		p.log.Error("no such pipeline", "requested", pp)
	}

	ppl := pipe.(*pipeline.Pipeline)

	d, ok := p.consumers[ppl.Name()]
	if !ok {
		p.log.Warn("driver for the pipeline not found", "pipeline", pp)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
	defer cancel()
	// redirect call to the underlying driver
	d.Resume(ctx, ppl.Name())
}

// Declare a pipeline.
func (p *Plugin) Declare(pipeline *pipeline.Pipeline) error {
	const op = errors.Op("jobs_plugin_declare")
	// driver for the pipeline (ie amqp, ephemeral, etc)
	dr := pipeline.Driver()
	if dr == "" {
		return errors.E(op, errors.Errorf("no associated driver with the pipeline, pipeline name: %s", pipeline.Name()))
	}

	// jobConstructors contains constructors for the drivers
	// we need here to initialize these drivers for the pipelines
	if c, ok := p.jobConstructors[dr]; ok {
		// init the driver from pipeline
		initializedDriver, err := c.FromPipeline(pipeline, p.events, p.queue)
		if err != nil {
			return errors.E(op, err)
		}

		// add driver to the set of the consumers (name - pipeline name, value - associated driver)
		p.consumers[pipeline.Name()] = initializedDriver

		// register pipeline for the initialized driver
		err = initializedDriver.Register(context.Background(), pipeline)
		if err != nil {
			return errors.E(op, errors.Errorf("pipe register failed for the driver: %s with pipe name: %s", pipeline.Driver(), pipeline.Name()))
		}

		// if pipeline initialized to be consumed, call Run on it
		// but likely for the dynamic pipelines it should be started manually
		if _, ok := p.consume[pipeline.Name()]; ok {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
			defer cancel()
			err = initializedDriver.Run(ctx, pipeline)
			if err != nil {
				return errors.E(op, err)
			}
		}
	}

	// save the pipeline
	p.pipelines.Store(pipeline.Name(), pipeline)

	return nil
}

// Destroy pipeline and release all associated resources.
func (p *Plugin) Destroy(pp string) error {
	const op = errors.Op("jobs_plugin_destroy")
	pipe, ok := p.pipelines.Load(pp)
	if !ok {
		return errors.E(op, errors.Errorf("no such pipeline, requested: %s", pp))
	}

	// type conversion
	ppl := pipe.(*pipeline.Pipeline)

	d, ok := p.consumers[ppl.Name()]
	if !ok {
		return errors.E(op, errors.Errorf("consumer not registered for the requested driver: %s", ppl.Driver()))
	}

	// delete consumer
	delete(p.consumers, ppl.Name())
	// delete old pipeline
	p.pipelines.LoadAndDelete(pp)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(p.cfg.Timeout))
	err := d.Stop(ctx)
	if err != nil {
		cancel()
		return errors.E(op, err)
	}

	d = nil
	cancel()
	return nil
}

func (p *Plugin) List() []string {
	out := make([]string, 0, 10)

	p.pipelines.Range(func(key, _ interface{}) bool {
		// we can safely convert value here as we know that we store keys as strings
		out = append(out, key.(string))
		return true
	})

	return out
}

func (p *Plugin) RPC() interface{} {
	return &rpc{
		log: p.log,
		p:   p,
	}
}

func (p *Plugin) collectJobsEvents(event interface{}) {
	if jev, ok := event.(events.JobEvent); ok {
		switch jev.Event {
		case events.EventPipePaused:
			p.log.Info("pipeline paused", "pipeline", jev.Pipeline, "driver", jev.Driver, "start", jev.Start.UTC(), "elapsed", jev.Elapsed)
		case events.EventJobStart:
			p.log.Info("job processing started", "start", jev.Start.UTC(), "elapsed", jev.Elapsed)
		case events.EventJobOK:
			p.log.Info("job processed without errors", "ID", jev.ID, "start", jev.Start.UTC(), "elapsed", jev.Elapsed)
		case events.EventPushOK:
			p.log.Info("job pushed to the queue", "start", jev.Start.UTC(), "elapsed", jev.Elapsed)
		case events.EventPushError:
			p.log.Error("job push error, job might be lost", "error", jev.Error, "pipeline", jev.Pipeline, "ID", jev.ID, "driver", jev.Driver, "start", jev.Start.UTC(), "elapsed", jev.Elapsed)
		case events.EventJobError:
			p.log.Error("job processed with errors", "error", jev.Error, "ID", jev.ID, "start", jev.Start.UTC(), "elapsed", jev.Elapsed)
		case events.EventPipeActive:
			p.log.Info("pipeline active", "pipeline", jev.Pipeline, "start", jev.Start.UTC(), "elapsed", jev.Elapsed)
		case events.EventPipeStopped:
			p.log.Warn("pipeline stopped", "pipeline", jev.Pipeline, "start", jev.Start.UTC(), "elapsed", jev.Elapsed)
		case events.EventPipeError:
			p.log.Error("pipeline error", "pipeline", jev.Pipeline, "error", jev.Error, "start", jev.Start.UTC(), "elapsed", jev.Elapsed)
		case events.EventDriverReady:
			p.log.Info("driver ready", "pipeline", jev.Pipeline, "start", jev.Start.UTC(), "elapsed", jev.Elapsed)
		}
	}
}

func (p *Plugin) getPayload(body, context []byte) *payload.Payload {
	pld := p.pldPool.Get().(*payload.Payload)
	pld.Body = body
	pld.Context = context
	return pld
}

func (p *Plugin) putPayload(pld *payload.Payload) {
	pld.Body = nil
	pld.Context = nil
	p.pldPool.Put(pld)
}
