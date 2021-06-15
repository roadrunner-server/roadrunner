package jobs

import (
	"fmt"
	//"github.com/sirupsen/logrus"
	//"github.com/spiral/roadrunner"
	//"github.com/spiral/roadrunner/service"
	//"github.com/spiral/roadrunner/service/env"
	//"github.com/spiral/roadrunner/service/rpc"
	"sync"
	"sync/atomic"
	"time"
)

// ID defines public service name.
const ID = "jobs"

// Service wraps roadrunner container and manage set of parent within it.
type Service struct {
	// Associated parent
	Brokers map[string]Broker

	// brokers and routing config
	cfg *Config

	// environment, logger and listeners
	//env env.Environment
	//log *logrus.Logger
	lsn []func(event int, ctx interface{})

	// server and server controller
	//rr *roadrunner.Server
	//cr roadrunner.Controller

	// task balancer
	execPool chan Handler

	// registered brokers
	serving int32
	//brokers service.Container

	// pipelines pipelines
	mup       sync.Mutex
	pipelines map[*Pipeline]bool
}

// Attach attaches cr. Currently only one cr is supported.
func (svc *Service) Attach(ctr roadrunner.Controller) {
	svc.cr = ctr
}

// AddListener attaches event listeners to the service and all underlying brokers.
func (svc *Service) AddListener(l func(event int, ctx interface{})) {
	svc.lsn = append(svc.lsn, l)
}

// Init configures job service.
func (svc *Service) Init(
	cfg service.Config,
	log *logrus.Logger,
	env env.Environment,
	rpc *rpc.Service,
) (ok bool, err error) {
	svc.cfg = &Config{}
	if err := svc.cfg.Hydrate(cfg); err != nil {
		return false, err
	}

	svc.env = env
	svc.log = log

	if rpc != nil {
		if err := rpc.Register(ID, &rpcServer{svc}); err != nil {
			return false, err
		}
	}

	// limit the number of parallel threads
	if svc.cfg.Workers.Command != "" {
		svc.execPool = make(chan Handler, svc.cfg.Workers.Pool.NumWorkers)
		for i := int64(0); i < svc.cfg.Workers.Pool.NumWorkers; i++ {
			svc.execPool <- svc.exec
		}

		svc.rr = roadrunner.NewServer(svc.cfg.Workers)
	}

	svc.pipelines = make(map[*Pipeline]bool)
	for _, p := range svc.cfg.pipelines {
		svc.pipelines[p] = false
	}

	// run all brokers in nested container
	svc.brokers = service.NewContainer(log)
	for name, b := range svc.Brokers {
		svc.brokers.Register(name, b)
		if ep, ok := b.(EventProvider); ok {
			ep.Listen(svc.throw)
		}
	}

	// init all broker configs
	if err := svc.brokers.Init(svc.cfg); err != nil {
		return false, err
	}

	// register all pipelines (per broker)
	for name, b := range svc.Brokers {
		for _, pipe := range svc.cfg.pipelines.Broker(name) {
			if err := b.Register(pipe); err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

// Serve serves local rr server and creates broker association.
func (svc *Service) Serve() error {
	if svc.rr != nil {
		if svc.env != nil {
			if err := svc.env.Copy(svc.cfg.Workers); err != nil {
				return err
			}
		}

		// ensure that workers aware of running within jobs
		svc.cfg.Workers.SetEnv("rr_jobs", "true")
		svc.rr.Listen(svc.throw)

		if svc.cr != nil {
			svc.rr.Attach(svc.cr)
		}

		if err := svc.rr.Start(); err != nil {
			return err
		}
		defer svc.rr.Stop()

		// start pipelines of all the pipelines
		for _, p := range svc.cfg.pipelines.Names(svc.cfg.Consume...) {
			// start pipeline consuming
			if err := svc.Consume(p, svc.execPool, svc.error); err != nil {
				svc.Stop()

				return err
			}
		}
	}

	atomic.StoreInt32(&svc.serving, 1)
	defer atomic.StoreInt32(&svc.serving, 0)

	return svc.brokers.Serve()
}

// Stop all pipelines and rr server.
func (svc *Service) Stop() {
	if atomic.LoadInt32(&svc.serving) == 0 {
		return
	}

	wg := sync.WaitGroup{}
	for _, p := range svc.cfg.pipelines.Names(svc.cfg.Consume...).Reverse() {
		wg.Add(1)

		go func(p *Pipeline) {
			defer wg.Done()
			if err := svc.Consume(p, nil, nil); err != nil {
				svc.throw(EventPipeError, &PipelineError{Pipeline: p, Caused: err})
			}
		}(p)
	}

	wg.Wait()
	svc.brokers.Stop()
}

// Server returns associated rr server (if any).
func (svc *Service) Server() *roadrunner.Server {
	return svc.rr
}

// Stat returns list of pipelines workers and their stats.
func (svc *Service) Stat(pipe *Pipeline) (stat *Stat, err error) {
	b, ok := svc.Brokers[pipe.Broker()]
	if !ok {
		return nil, fmt.Errorf("undefined broker `%s`", pipe.Broker())
	}

	stat, err = b.Stat(pipe)
	if err != nil {
		return nil, err
	}

	stat.Pipeline = pipe.Name()
	stat.Broker = pipe.Broker()

	svc.mup.Lock()
	stat.Consuming = svc.pipelines[pipe]
	svc.mup.Unlock()

	return stat, err
}

// Consume enables or disables pipeline pipelines using given handlers.
func (svc *Service) Consume(pipe *Pipeline, execPool chan Handler, errHandler ErrorHandler) error {
	svc.mup.Lock()

	if execPool != nil {
		if svc.pipelines[pipe] {
			svc.mup.Unlock()
			return nil
		}

		svc.throw(EventPipeConsume, pipe)
		svc.pipelines[pipe] = true
	} else {
		if !svc.pipelines[pipe] {
			svc.mup.Unlock()
			return nil
		}

		svc.throw(EventPipeStop, pipe)
		svc.pipelines[pipe] = false
	}

	broker, ok := svc.Brokers[pipe.Broker()]
	if !ok {
		svc.mup.Unlock()
		return fmt.Errorf("undefined broker `%s`", pipe.Broker())
	}
	svc.mup.Unlock()

	if err := broker.Consume(pipe, execPool, errHandler); err != nil {
		svc.mup.Lock()
		svc.pipelines[pipe] = false
		svc.mup.Unlock()

		svc.throw(EventPipeError, &PipelineError{Pipeline: pipe, Caused: err})

		return err
	}

	if execPool != nil {
		svc.throw(EventPipeActive, pipe)
	} else {
		svc.throw(EventPipeStopped, pipe)
	}

	return nil
}

// Push job to associated broker and return job id.
func (svc *Service) Push(job *Job) (string, error) {
	pipe, pOpts, err := svc.cfg.MatchPipeline(job)
	if err != nil {
		return "", err
	}

	if pOpts != nil {
		job.Options.Merge(pOpts)
	}

	broker, ok := svc.Brokers[pipe.Broker()]
	if !ok {
		return "", fmt.Errorf("undefined broker `%s`", pipe.Broker())
	}

	id, err := broker.Push(pipe, job)

	if err != nil {
		svc.throw(EventPushError, &JobError{Job: job, Caused: err})
	} else {
		svc.throw(EventPushOK, &JobEvent{ID: id, Job: job})
	}

	return id, err
}

// exec executed job using local RR server. Make sure that service is started.
func (svc *Service) exec(id string, j *Job) error {
	start := time.Now()
	svc.throw(EventJobStart, &JobEvent{ID: id, Job: j, start: start})

	// ignore response for now, possibly add more routing options
	_, err := svc.rr.Exec(&roadrunner.Payload{
		Body:    j.Body(),
		Context: j.Context(id),
	})

	if err == nil {
		svc.throw(EventJobOK, &JobEvent{
			ID:      id,
			Job:     j,
			start:   start,
			elapsed: time.Since(start),
		})
	} else {
		svc.throw(EventJobError, &JobError{
			ID:     id,
			Job:    j,
			Caused: err, start: start,
			elapsed: time.Since(start),
		})
	}

	return err
}

// register died job, can be used to move to fallback testQueue or log
func (svc *Service) error(id string, j *Job, err error) {
	// nothing for now, possibly route to another pipeline
}

// throw handles service, server and pool events.
func (svc *Service) throw(event int, ctx interface{}) {
	for _, l := range svc.lsn {
		l(event, ctx)
	}

	if event == roadrunner.EventServerFailure {
		// underlying rr server is dead, stop everything
		svc.Stop()
	}
}
