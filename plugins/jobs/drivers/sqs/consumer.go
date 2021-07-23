package sqs

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	cfgPlugin "github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/job"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type JobConsumer struct {
	sync.Mutex
	pq       priorityqueue.Queue
	log      logger.Logger
	eh       events.Handler
	pipeline atomic.Value

	// connection info
	key               string
	secret            string
	sessionToken      string
	region            string
	endpoint          string
	queue             *string
	messageGroupID    string
	waitTime          int32
	prefetch          int32
	visibilityTimeout int32

	// if user invoke several resume operations
	listeners uint32

	// queue optional parameters
	attributes map[string]string
	tags       map[string]string

	client   *sqs.Client
	queueURL *string

	pauseCh chan struct{}
}

func NewSQSConsumer(configKey string, log logger.Logger, cfg cfgPlugin.Configurer, e events.Handler, pq priorityqueue.Queue) (*JobConsumer, error) {
	const op = errors.Op("new_sqs_consumer")

	// if no such key - error
	if !cfg.Has(configKey) {
		return nil, errors.E(op, errors.Errorf("no configuration by provided key: %s", configKey))
	}

	// if no global section
	if !cfg.Has(pluginName) {
		return nil, errors.E(op, errors.Str("no global sqs configuration, global configuration should contain sqs section"))
	}

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
	jb := &JobConsumer{
		pq:                pq,
		log:               log,
		eh:                e,
		messageGroupID:    uuid.NewString(),
		attributes:        pipeCfg.Attributes,
		tags:              pipeCfg.Tags,
		queue:             pipeCfg.Queue,
		prefetch:          pipeCfg.Prefetch,
		visibilityTimeout: pipeCfg.VisibilityTimeout,
		waitTime:          pipeCfg.WaitTimeSeconds,
		region:            globalCfg.Region,
		key:               globalCfg.Key,
		sessionToken:      globalCfg.SessionToken,
		secret:            globalCfg.Secret,
		endpoint:          globalCfg.Endpoint,
		pauseCh:           make(chan struct{}, 1),
	}

	// PARSE CONFIGURATION -------

	awsConf, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(globalCfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(jb.key, jb.secret, jb.sessionToken)))
	if err != nil {
		return nil, errors.E(op, err)
	}

	// config with retries
	jb.client = sqs.NewFromConfig(awsConf, sqs.WithEndpointResolver(sqs.EndpointResolverFromURL(jb.endpoint)), func(o *sqs.Options) {
		o.Retryer = retry.NewStandard(func(opts *retry.StandardOptions) {
			opts.MaxAttempts = 60
		})
	})

	out, err := jb.client.CreateQueue(context.Background(), &sqs.CreateQueueInput{QueueName: jb.queue, Attributes: jb.attributes, Tags: jb.tags})
	if err != nil {
		return nil, errors.E(op, err)
	}

	// assign a queue URL
	jb.queueURL = out.QueueUrl

	// To successfully create a new queue, you must provide a
	// queue name that adheres to the limits related to queues
	// (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/limits-queues.html)
	// and is unique within the scope of your queues. After you create a queue, you
	// must wait at least one second after the queue is created to be able to use the <------------
	// queue. To get the queue URL, use the GetQueueUrl action. GetQueueUrl require
	time.Sleep(time.Second * 2)

	return jb, nil
}

func FromPipeline(pipe *pipeline.Pipeline, log logger.Logger, cfg cfgPlugin.Configurer, e events.Handler, pq priorityqueue.Queue) (*JobConsumer, error) {
	const op = errors.Op("new_sqs_consumer")

	// if no global section
	if !cfg.Has(pluginName) {
		return nil, errors.E(op, errors.Str("no global sqs configuration, global configuration should contain sqs section"))
	}

	// PARSE CONFIGURATION -------
	var globalCfg GlobalCfg

	err := cfg.UnmarshalKey(pluginName, &globalCfg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	globalCfg.InitDefault()

	res := make(map[string]interface{})
	pipe.Map(attributes, res)

	attr := make(map[string]string)
	// accept only string values
	for i := range res {
		if v, ok := res[i].(string); ok {
			attr[i] = v
		}
	}

	// delete all values with map.clear to reuse for the tags
	for k := range res {
		delete(res, k)
	}

	pipe.Map(tags, res)

	tg := make(map[string]string)
	// accept only string values
	for i := range res {
		if v, ok := res[i].(string); ok {
			attr[i] = v
		}
	}

	// initialize job consumer
	jb := &JobConsumer{
		pq:                pq,
		log:               log,
		eh:                e,
		messageGroupID:    uuid.NewString(),
		attributes:        attr,
		tags:              tg,
		queue:             aws.String(pipe.String(queue, "default")),
		prefetch:          int32(pipe.Int(pref, 10)),
		visibilityTimeout: int32(pipe.Int(visibility, 0)),
		waitTime:          int32(pipe.Int(waitTime, 0)),
		region:            globalCfg.Region,
		key:               globalCfg.Key,
		sessionToken:      globalCfg.SessionToken,
		secret:            globalCfg.Secret,
		endpoint:          globalCfg.Endpoint,
		pauseCh:           make(chan struct{}, 1),
	}

	// PARSE CONFIGURATION -------

	awsConf, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(globalCfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(jb.key, jb.secret, jb.sessionToken)))
	if err != nil {
		return nil, errors.E(op, err)
	}

	// config with retries
	jb.client = sqs.NewFromConfig(awsConf, sqs.WithEndpointResolver(sqs.EndpointResolverFromURL(jb.endpoint)), func(o *sqs.Options) {
		o.Retryer = retry.NewStandard(func(opts *retry.StandardOptions) {
			opts.MaxAttempts = 60
		})
	})

	out, err := jb.client.CreateQueue(context.Background(), &sqs.CreateQueueInput{QueueName: jb.queue, Attributes: jb.attributes, Tags: jb.tags})
	if err != nil {
		return nil, errors.E(op, err)
	}

	// assign a queue URL
	jb.queueURL = out.QueueUrl

	// To successfully create a new queue, you must provide a
	// queue name that adheres to the limits related to queues
	// (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/limits-queues.html)
	// and is unique within the scope of your queues. After you create a queue, you
	// must wait at least one second after the queue is created to be able to use the <------------
	// queue. To get the queue URL, use the GetQueueUrl action. GetQueueUrl require
	time.Sleep(time.Second * 2)

	return jb, nil
}

func (j *JobConsumer) Push(jb *job.Job) error {
	const op = errors.Op("sqs_push")
	// check if the pipeline registered

	// load atomic value
	pipe := j.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != jb.Options.Pipeline {
		return errors.E(op, errors.Errorf("no such pipeline: %s, actual: %s", jb.Options.Pipeline, pipe.Name()))
	}

	// The length of time, in seconds, for which to delay a specific message. Valid
	// values: 0 to 900. Maximum: 15 minutes.
	if jb.Options.Delay > 900 {
		return errors.E(op, errors.Errorf("unable to push, maximum possible delay is 900 seconds (15 minutes), provided: %d", jb.Options.Delay))
	}

	msg := fromJob(jb)

	// 10 seconds deadline to make a request TODO ???
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
	defer cancel()
	// The new value for the message's visibility timeout (in seconds). Values range: 0
	// to 43200. Maximum: 12 hours.
	_, err := j.client.SendMessage(ctx, msg.pack(j.queueURL))
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (j *JobConsumer) Register(pipeline *pipeline.Pipeline) error {
	j.pipeline.Store(pipeline)
	return nil
}

func (j *JobConsumer) Run(p *pipeline.Pipeline) error {
	const op = errors.Op("rabbit_consume")

	j.Lock()
	defer j.Unlock()

	pipe := j.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p.Name() {
		return errors.E(op, errors.Errorf("no such pipeline registered: %s", pipe.Name()))
	}

	atomic.AddUint32(&j.listeners, 1)

	// start listener
	go j.listen()

	j.eh.Push(events.JobEvent{
		Event:    events.EventPipeActive,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    time.Now(),
	})

	return nil
}

func (j *JobConsumer) Stop() error {
	j.pauseCh <- struct{}{}

	pipe := j.pipeline.Load().(*pipeline.Pipeline)
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

	// stop consume
	j.pauseCh <- struct{}{}

	j.eh.Push(events.JobEvent{
		Event:    events.EventPipePaused,
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
