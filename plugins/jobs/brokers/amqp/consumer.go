package amqp

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/common/jobs"
	"github.com/spiral/roadrunner/v2/pkg/priorityqueue"
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
	logger logger.Logger
	pq     priorityqueue.Queue

	pipelines sync.Map

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

	// TODO send data to channel
	stop chan struct{}
}

func NewAMQPConsumer(configKey string, log logger.Logger, cfg config.Configurer, pq priorityqueue.Queue) (jobs.Consumer, error) {
	const op = errors.Op("new_amqp_consumer")
	// we need to obtain two parts of the amqp information here.
	// firs part - address to connect, it is located in the global section under the amqp pluginName
	// second part - queues and other pipeline information
	jb := &JobsConsumer{
		logger:       log,
		pq:           pq,
		consumeID:    uuid.NewString(),
		stop:         make(chan struct{}),
		retryTimeout: time.Minute * 5,
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

	err = cfg.UnmarshalKey(configKey, &globalCfg)
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

	// assign address
	jb.connStr = globalCfg.Addr

	err = jb.initRabbitMQ()
	if err != nil {
		return nil, err
	}

	jb.publishChan, err = jb.conn.Channel()
	if err != nil {
		panic(err)
	}

	// run redialer for the connection
	jb.redialer()

	return jb, nil
}

func (j *JobsConsumer) Push(job *structs.Job) error {
	const op = errors.Op("ephemeral_push")
	// lock needed here to re-create a connections and channels in case of error
	j.RLock()
	defer j.RUnlock()

	// convert
	msg := FromJob(job)

	// check if the pipeline registered
	if _, ok := j.pipelines.Load(job.Options.Pipeline); ok {
		// handle timeouts
		if job.Options.DelayDuration() > 0 {
			// TODO declare separate method for this if condition

			delayMs := int64(job.Options.DelayDuration().Seconds() * 1000)
			tmpQ := fmt.Sprintf("delayed-%d.%s.%s", delayMs, j.exchangeName, j.queue)

			_, err := j.publishChan.QueueDeclare(tmpQ, true, false, false, false, amqp.Table{
				dlx:           j.exchangeName,
				dlxRoutingKey: j.routingKey,
				dlxTTL:        delayMs,
				dlxExpires:    delayMs * 2,
			})

			if err != nil {
				panic(err)
			}

			err = j.publishChan.QueueBind(tmpQ, tmpQ, j.exchangeName, false, nil)
			if err != nil {
				panic(err)
			}

			// insert to the local, limited pipeline
			err = j.publishChan.Publish(j.exchangeName, tmpQ, false, false, amqp.Publishing{
				Headers:     pack(job.Ident, 0, msg),
				ContentType: contentType,
				Timestamp:   time.Now(),
				Body:        nil,
			})
			if err != nil {
				panic(err)
			}

			return nil
		}

		// insert to the local, limited pipeline
		err := j.publishChan.Publish(j.exchangeName, j.routingKey, false, false, amqp.Publishing{
			Headers:     pack(job.Ident, 0, msg),
			ContentType: contentType,
			Timestamp:   time.Now(),
			Body:        nil,
		})
		if err != nil {
			panic(err)
		}

		return nil
	}

	return errors.E(op, errors.Errorf("no such pipeline: %s", job.Options.Pipeline))
}

func (j *JobsConsumer) Register(pipeline *pipeline.Pipeline) error {
	const op = errors.Op("rabbitmq_register")
	if _, ok := j.pipelines.Load(pipeline.Name()); ok {
		return errors.E(op, errors.Errorf("queue %s has already been registered", pipeline))
	}

	j.pipelines.Store(pipeline.Name(), true)

	return nil
}

func (j *JobsConsumer) Consume(pipeline *pipeline.Pipeline) error {
	const op = errors.Op("rabbit_consume")

	if _, ok := j.pipelines.Load(pipeline.Name()); !ok {
		return errors.E(op, errors.Errorf("no such pipeline registered: %s", pipeline.Name()))
	}

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

func (j *JobsConsumer) List() []*pipeline.Pipeline {
	panic("implement me")
}

func (j *JobsConsumer) Pause(pipeline string) {
	if q, ok := j.pipelines.Load(pipeline); ok {
		if q == true {
			// mark pipeline as turned off
			j.pipelines.Store(pipeline, false)
		}
	}

	err := j.publishChan.Cancel(j.consumeID, true)
	if err != nil {
		j.logger.Error("cancel publish channel, forcing close", "error", err)
		errCl := j.publishChan.Close()
		if errCl != nil {
			j.logger.Error("force close failed", "error", err)
		}
	}
}

func (j *JobsConsumer) Resume(pipeline string) {
	if q, ok := j.pipelines.Load(pipeline); ok {
		if q == false {
			// mark pipeline as turned off
			j.pipelines.Store(pipeline, true)
		}
		var err error
		j.consumeChan, err = j.conn.Channel()
		if err != nil {
			j.logger.Error("create channel on rabbitmq connection", "error", err)
			return
		}

		err = j.consumeChan.Qos(j.prefetchCount, 0, false)
		if err != nil {
			j.logger.Error("qos set failed", "error", err)
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
			j.logger.Error("consume operation failed", "error", err)
			return
		}

		// run listener
		j.listener(deliv)
	}
}

// Declare used to dynamically declare a pipeline
func (j *JobsConsumer) Declare(pipeline *pipeline.Pipeline) error {
	pipeline.String(exchangeKey, "")
	pipeline.String(queue, "")
	pipeline.String(routingKey, "")
	pipeline.String(exchangeType, "direct")
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
