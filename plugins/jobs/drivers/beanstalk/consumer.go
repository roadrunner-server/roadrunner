package beanstalk

import (
	"strings"
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
	log logger.Logger
	eh  events.Handler
	pq  priorityqueue.Queue

	// beanstalk
	conn *beanstalk.Conn
	tout time.Duration
	// tube name
	tName string
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

	// initialize job consumer
	jc := &JobConsumer{
		pq:   pq,
		log:  log,
		eh:   e,
		tout: globalCfg.Timeout,
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

func FromPipeline(pipe *pipeline.Pipeline, log logger.Logger, cfg config.Configurer, e events.Handler, pq priorityqueue.Queue) (*JobConsumer, error) {
	const op = errors.Op("new_beanstalk_consumer")

	const (
		tube string = "tube"
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
		pq:    pq,
		log:   log,
		eh:    e,
		tout:  globalCfg.Timeout,
		tName: pipe.String(tube, ""),
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
func (j *JobConsumer) Push(job *job.Job) error {
	panic("implement me")
}

func (j *JobConsumer) Register(pipeline *pipeline.Pipeline) error {
	panic("implement me")
}

func (j *JobConsumer) Run(pipeline *pipeline.Pipeline) error {
	panic("implement me")
}

func (j *JobConsumer) Stop() error {
	panic("implement me")
}

func (j *JobConsumer) Pause(pipeline string) {
	panic("implement me")
}

func (j *JobConsumer) Resume(pipeline string) {
	panic("implement me")
}
