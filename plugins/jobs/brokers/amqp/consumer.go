package amqp

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/streadway/amqp"
)

// pipeline rabbitmq info
const (
	exchangeKey  string = "exchange"
	exchangeType string = "exchange-type"
	queue        string = "queue"
	routingKey   string = "routing-key"
	prefetch     string = "prefetch"

	dlx           string = "x-dead-letter-exchange"
	dlxRoutingKey string = "x-dead-letter-routing-key"
	dlxTTL        string = "x-message-ttl"
	dlxExpires    string = "x-expires"

	contentType string = "application/octet-stream"
)

type GlobalCfg struct {
	Addr string `mapstructure:"addr"`
}

// Config is used to parse pipeline configuration
type Config struct {
	PrefetchCount int    `mapstructure:"pipeline_size"`
	Queue         string `mapstructure:"queue"`
	Exchange      string `mapstructure:"exchange"`
	ExchangeType  string `mapstructure:"exchange_type"`
	RoutingKey    string `mapstructure:"routing_key"`
}

type JobsConsumer struct {
	sync.RWMutex
	log logger.Logger
	pq  priorityqueue.Queue
	eh  events.Handler

	pipeline atomic.Value

	// amqp connection
	conn        *amqp.Connection
	consumeChan *amqp.Channel
	publishChan *amqp.Channel

	retryTimeout  time.Duration
	prefetchCount int
	exchangeName  string
	queue         string
	consumeID     string
	connStr       string
	exchangeType  string
	routingKey    string

	delayCache map[string]struct{}

	stopCh chan struct{}
}

// NewAMQPConsumer initializes rabbitmq pipeline
func NewAMQPConsumer(configKey string, log logger.Logger, cfg config.Configurer, e events.Handler, pq priorityqueue.Queue) (*JobsConsumer, error) {
	const op = errors.Op("new_amqp_consumer")
	// we need to obtain two parts of the amqp information here.
	// firs part - address to connect, it is located in the global section under the amqp pluginName
	// second part - queues and other pipeline information
	jb := &JobsConsumer{
		log:       log,
		pq:        pq,
		eh:        e,
		consumeID: uuid.NewString(),
		stopCh:    make(chan struct{}),
		// TODO to config
		retryTimeout: time.Minute * 5,
		delayCache:   make(map[string]struct{}, 100),
	}

	// if no such key - error
	if !cfg.Has(configKey) {
		return nil, errors.E(op, errors.Errorf("no configuration by provided key: %s", configKey))
	}

	// if no global section
	if !cfg.Has(pluginName) {
		return nil, errors.E(op, errors.Str("no global amqp configuration, global configuration should contain amqp addrs"))
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

	jb.routingKey = pipeCfg.RoutingKey
	jb.queue = pipeCfg.Queue
	jb.exchangeType = pipeCfg.ExchangeType
	jb.exchangeName = pipeCfg.Exchange
	jb.prefetchCount = pipeCfg.PrefetchCount

	// PARSE CONFIGURATION -------

	jb.conn, err = amqp.Dial(globalCfg.Addr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// save address
	jb.connStr = globalCfg.Addr

	err = jb.initRabbitMQ()
	if err != nil {
		return nil, errors.E(op, err)
	}

	jb.publishChan, err = jb.conn.Channel()
	if err != nil {
		return nil, errors.E(op, err)
	}

	// run redialer for the connection
	jb.redialer()

	return jb, nil
}

func FromPipeline(pipeline *pipeline.Pipeline, log logger.Logger, cfg config.Configurer, e events.Handler, pq priorityqueue.Queue) (*JobsConsumer, error) {
	const op = errors.Op("new_amqp_consumer_from_pipeline")
	// we need to obtain two parts of the amqp information here.
	// firs part - address to connect, it is located in the global section under the amqp pluginName
	// second part - queues and other pipeline information
	jb := &JobsConsumer{
		log:          log,
		eh:           e,
		pq:           pq,
		consumeID:    uuid.NewString(),
		stopCh:       make(chan struct{}),
		retryTimeout: time.Minute * 5,
		delayCache:   make(map[string]struct{}, 100),
	}

	// if no global section
	if !cfg.Has(pluginName) {
		return nil, errors.E(op, errors.Str("no global amqp configuration, global configuration should contain amqp addrs"))
	}

	// PARSE CONFIGURATION -------
	var globalCfg GlobalCfg

	err := cfg.UnmarshalKey(pluginName, &globalCfg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	globalCfg.InitDefault()

	jb.routingKey = pipeline.String(routingKey, "")
	jb.queue = pipeline.String(queue, "default")
	jb.exchangeType = pipeline.String(exchangeType, "direct")
	jb.exchangeName = pipeline.String(exchangeKey, "amqp.default")
	jb.prefetchCount = pipeline.Int(prefetch, 10)

	// PARSE CONFIGURATION -------

	jb.conn, err = amqp.Dial(globalCfg.Addr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// save address
	jb.connStr = globalCfg.Addr

	err = jb.initRabbitMQ()
	if err != nil {
		return nil, errors.E(op, err)
	}

	jb.publishChan, err = jb.conn.Channel()
	if err != nil {
		return nil, errors.E(op, err)
	}

	// run redialer for the connection
	jb.redialer()

	return jb, nil
}

func (j *JobsConsumer) Push(job *structs.Job) error {
	const op = errors.Op("rabbitmq_push")
	// check if the pipeline registered

	// load atomic value
	pipe := j.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != job.Options.Pipeline {
		return errors.E(op, errors.Errorf("no such pipeline: %s, actual: %s", job.Options.Pipeline, pipe.Name()))
	}

	// lock needed here to protect redial concurrent operation
	// we may be in the redial state here
	j.Lock()
	defer j.Unlock()

	// convert
	msg := fromJob(job)
	p, err := pack(job.Ident, msg)
	if err != nil {
		return errors.E(op, err)
	}

	// handle timeouts
	if job.Options.DelayDuration() > 0 {
		// TODO declare separate method for this if condition
		delayMs := int64(job.Options.DelayDuration().Seconds() * 1000)
		tmpQ := fmt.Sprintf("delayed-%d.%s.%s", delayMs, j.exchangeName, j.queue)

		// delay cache optimization.
		// If user already declared a queue with a delay, do not redeclare and rebind the queue
		// Before -> 2.5k RPS with redeclaration
		// After -> 30k RPS
		if _, exists := j.delayCache[tmpQ]; exists {
			// insert to the local, limited pipeline
			err = j.publishChan.Publish(j.exchangeName, tmpQ, false, false, amqp.Publishing{
				Headers:      p,
				ContentType:  contentType,
				Timestamp:    time.Now(),
				DeliveryMode: amqp.Persistent,
				Body:         msg.Body(),
			})

			if err != nil {
				return errors.E(op, err)
			}

			return nil
		}

		_, err = j.publishChan.QueueDeclare(tmpQ, true, false, false, false, amqp.Table{
			dlx:           j.exchangeName,
			dlxRoutingKey: j.routingKey,
			dlxTTL:        delayMs,
			dlxExpires:    delayMs * 2,
		})

		if err != nil {
			return errors.E(op, err)
		}

		err = j.publishChan.QueueBind(tmpQ, tmpQ, j.exchangeName, false, nil)
		if err != nil {
			return errors.E(op, err)
		}

		// insert to the local, limited pipeline
		err = j.publishChan.Publish(j.exchangeName, tmpQ, false, false, amqp.Publishing{
			Headers:      p,
			ContentType:  contentType,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent,
			Body:         msg.Body(),
		})

		if err != nil {
			return errors.E(op, err)
		}

		j.delayCache[tmpQ] = struct{}{}

		return nil
	}

	// insert to the local, limited pipeline
	err = j.publishChan.Publish(j.exchangeName, j.routingKey, false, false, amqp.Publishing{
		Headers:      p,
		ContentType:  contentType,
		Timestamp:    time.Now(),
		DeliveryMode: amqp.Persistent,
		Body:         msg.Body(),
	})
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (j *JobsConsumer) Register(pipeline *pipeline.Pipeline) error {
	j.pipeline.Store(pipeline)
	return nil
}

func (j *JobsConsumer) Run(p *pipeline.Pipeline) error {
	const op = errors.Op("rabbit_consume")

	pipe := j.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p.Name() {
		return errors.E(op, errors.Errorf("no such pipeline registered: %s", pipe.Name()))
	}

	// protect connection (redial)
	j.Lock()
	defer j.Unlock()

	var err error
	j.consumeChan, err = j.conn.Channel()
	if err != nil {
		return errors.E(op, err)
	}

	err = j.consumeChan.Qos(j.prefetchCount, 0, false)
	if err != nil {
		return errors.E(op, err)
	}

	// start reading messages from the channel
	deliv, err := j.consumeChan.Consume(
		j.queue,
		j.consumeID,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.E(op, err)
	}

	// run listener
	j.listener(deliv)

	return nil
}

func (j *JobsConsumer) Pause(p string) {
	pipe := j.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		j.log.Error("no such pipeline", "requested pause on: ", p)
	}

	// protect connection (redial)
	j.Lock()
	defer j.Unlock()

	err := j.consumeChan.Cancel(j.consumeID, true)
	if err != nil {
		j.log.Error("cancel publish channel, forcing close", "error", err)
		errCl := j.consumeChan.Close()
		if errCl != nil {
			j.log.Error("force close failed", "error", err)
		}
	}
}

func (j *JobsConsumer) Resume(p string) {
	pipe := j.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		j.log.Error("no such pipeline", "requested resume on: ", p)
	}

	// protect connection (redial)
	j.Lock()
	defer j.Unlock()

	var err error
	j.consumeChan, err = j.conn.Channel()
	if err != nil {
		j.log.Error("create channel on rabbitmq connection", "error", err)
		return
	}

	err = j.consumeChan.Qos(j.prefetchCount, 0, false)
	if err != nil {
		j.log.Error("qos set failed", "error", err)
		return
	}

	// start reading messages from the channel
	deliv, err := j.consumeChan.Consume(
		j.queue,
		j.consumeID,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		j.log.Error("consume operation failed", "error", err)
		return
	}

	// run listener
	j.listener(deliv)
}

func (j *JobsConsumer) Stop() error {
	j.stopCh <- struct{}{}

	pipe := j.pipeline.Load().(*pipeline.Pipeline)
	j.eh.Push(events.JobEvent{
		Event:    events.EventPipeStopped,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    time.Now(),
		Elapsed:  0,
	})
	return nil
}

func (c *Config) InitDefault() {
	if c.ExchangeType == "" {
		c.ExchangeType = "direct"
	}

	if c.Exchange == "" {
		c.Exchange = "default"
	}

	if c.PrefetchCount == 0 {
		c.PrefetchCount = 100
	}
}

func (c *GlobalCfg) InitDefault() {
	if c.Addr == "" {
		c.Addr = "amqp://guest:guest@localhost:5672/"
	}
}
