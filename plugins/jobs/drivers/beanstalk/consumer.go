package beanstalk

import (
	"bytes"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/beanstalkd/go-beanstalk"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/job"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type JobConsumer struct {
	sync.Mutex
	log logger.Logger
	eh  events.Handler
	pq  priorityqueue.Queue

	pipeline  atomic.Value
	listeners uint32

	// beanstalk
	addr           string
	network        string
	conn           *beanstalk.Conn
	tube           *beanstalk.Tube
	tubeSet        *beanstalk.TubeSet
	reserveTimeout time.Duration
	reconnectCh    chan struct{}
	tout           time.Duration
	// tube name
	tName        string
	tubePriority uint32
	priority     int64

	stopCh chan struct{}
}

func NewBeanstalkConsumer(configKey string, log logger.Logger, cfg config.Configurer, e events.Handler, pq priorityqueue.Queue) (*JobConsumer, error) {
	const op = errors.Op("new_beanstalk_consumer")

	// PARSE CONFIGURATION -------
	var pipeCfg Config
	var globalCfg GlobalCfg

	err := cfg.UnmarshalKey(configKey, &pipeCfg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	pipeCfg.InitDefault()

	err = cfg.UnmarshalKey(pluginName, &globalCfg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	globalCfg.InitDefault()

	// PARSE CONFIGURATION -------

	dsn := strings.Split(globalCfg.Addr, "://")
	if len(dsn) != 2 {
		return nil, errors.E(op, errors.Errorf("invalid socket DSN (tcp://localhost:11300, unix://beanstalk.sock), provided: %s", globalCfg.Addr))
	}

	// initialize job consumer
	jc := &JobConsumer{
		pq:             pq,
		log:            log,
		eh:             e,
		network:        dsn[0],
		addr:           dsn[1],
		tout:           globalCfg.Timeout,
		reserveTimeout: pipeCfg.ReserveTimeout,
		tName:          pipeCfg.Tube,
		tubePriority:   pipeCfg.TubePriority,
		priority:       pipeCfg.PipePriority,

		// buffered with two because jobs root plugin can call Stop at the same time as Pause
		stopCh:      make(chan struct{}, 2),
		reconnectCh: make(chan struct{}),
	}

	jc.conn, err = beanstalk.DialTimeout(jc.network, jc.addr, jc.tout)
	if err != nil {
		return nil, err
	}

	jc.tube = beanstalk.NewTube(jc.conn, jc.tName)
	jc.tubeSet = beanstalk.NewTubeSet(jc.conn, jc.tName)

	// start redial listener
	go jc.redial()

	return jc, nil
}

func FromPipeline(pipe *pipeline.Pipeline, log logger.Logger, cfg config.Configurer, e events.Handler, pq priorityqueue.Queue) (*JobConsumer, error) {
	const op = errors.Op("new_beanstalk_consumer")

	const (
		tube           string = "tube"
		reserveTimeout string = "reserve_timeout"
	)

	// PARSE CONFIGURATION -------
	var globalCfg GlobalCfg

	err := cfg.UnmarshalKey(pluginName, &globalCfg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	globalCfg.InitDefault()

	// initialize job consumer
	jc := &JobConsumer{
		pq:             pq,
		log:            log,
		eh:             e,
		tout:           globalCfg.Timeout,
		tName:          pipe.String(tube, ""),
		reserveTimeout: time.Second * time.Duration(pipe.Int(reserveTimeout, 5)),
		stopCh:         make(chan struct{}),
		reconnectCh:    make(chan struct{}),
	}

	// PARSE CONFIGURATION -------

	dsn := strings.Split(globalCfg.Addr, "://")
	if len(dsn) != 2 {
		return nil, errors.E(op, errors.Errorf("invalid socket DSN (tcp://localhost:11300, unix://beanstalk.sock), provided: %s", globalCfg.Addr))
	}

	jc.conn, err = beanstalk.DialTimeout(dsn[0], dsn[1], jc.tout)
	if err != nil {
		return nil, err
	}

	// start redial listener
	go jc.redial()

	return jc, nil
}
func (j *JobConsumer) Push(jb *job.Job) error {
	const op = errors.Op("beanstalk_push")
	// check if the pipeline registered

	// load atomic value
	pipe := j.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != jb.Options.Pipeline {
		return errors.E(op, errors.Errorf("no such pipeline: %s, actual: %s", jb.Options.Pipeline, pipe.Name()))
	}

	// reconnect protection
	j.Lock()
	defer j.Unlock()

	item := fromJob(jb)

	bb := new(bytes.Buffer)
	bb.Grow(64)
	err := item.pack(bb)
	if err != nil {
		return errors.E(op, err)
	}

	// https://github.com/beanstalkd/beanstalkd/blob/master/doc/protocol.txt#L458
	// <pri> is an integer < 2**32. Jobs with smaller priority values will be
	// scheduled before jobs with larger priorities. The most urgent priority is 0;
	// the least urgent priority is 4,294,967,295.
	//
	// <delay> is an integer number of seconds to wait before putting the job in
	// the ready queue. The job will be in the "delayed" state during this time.
	// Maximum delay is 2**32-1.
	//
	// <ttr> -- time to run -- is an integer number of seconds to allow a worker
	// to run this job. This time is counted from the moment a worker reserves
	// this job. If the worker does not delete, release, or bury the job within
	// <ttr> seconds, the job will time out and the server will release the job.
	//	The minimum ttr is 1. If the client sends 0, the server will silently
	// increase the ttr to 1. Maximum ttr is 2**32-1.
	id, err := j.tube.Put(bb.Bytes(), 0, item.Options.DelayDuration(), item.Options.TimeoutDuration())
	if err != nil {
		errD := j.conn.Delete(id)
		if errD != nil {
			return errors.E(op, errors.Errorf("%s:%s", err.Error(), errD.Error()))
		}
		return errors.E(op, err)
	}

	return nil
}

func (j *JobConsumer) Register(pipeline *pipeline.Pipeline) error {
	// register the pipeline
	j.pipeline.Store(pipeline)
	return nil
}

func (j *JobConsumer) Run(p *pipeline.Pipeline) error {
	const op = errors.Op("beanstalk_run")
	// check if the pipeline registered

	// load atomic value
	pipe := j.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p.Name() {
		return errors.E(op, errors.Errorf("no such pipeline: %s, actual: %s", p.Name(), pipe.Name()))
	}

	atomic.AddUint32(&j.listeners, 1)

	j.Lock()
	defer j.Unlock()

	go j.listen()

	return nil
}

func (j *JobConsumer) Stop() error {
	pipe := j.pipeline.Load().(*pipeline.Pipeline)

	j.stopCh <- struct{}{}

	j.eh.Push(events.JobEvent{
		Event:    events.EventPipeStopped,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    time.Now(),
	})

	return nil
}

func (j *JobConsumer) Pause(p string) {
	// load atomic value
	pipe := j.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		j.log.Error("no such pipeline", "requested", p, "actual", pipe.Name())
		return
	}

	l := atomic.LoadUint32(&j.listeners)
	// no active listeners
	if l == 0 {
		j.log.Warn("no active listeners, nothing to pause")
		return
	}

	atomic.AddUint32(&j.listeners, ^uint32(0))

	j.stopCh <- struct{}{}

	j.eh.Push(events.JobEvent{
		Event:    events.EventPipeStopped,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    time.Now(),
	})
}

func (j *JobConsumer) Resume(p string) {
	// load atomic value
	pipe := j.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		j.log.Error("no such pipeline", "requested", p, "actual", pipe.Name())
		return
	}

	l := atomic.LoadUint32(&j.listeners)
	// no active listeners
	if l == 1 {
		j.log.Warn("sqs listener already in the active state")
		return
	}

	// start listener
	go j.listen()

	// increase num of listeners
	atomic.AddUint32(&j.listeners, 1)

	j.eh.Push(events.JobEvent{
		Event:    events.EventPipeActive,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    time.Now(),
	})
}