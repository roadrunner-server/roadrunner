package amqp

import (
	"errors"
	"fmt"
	"github.com/spiral/jobs/v2"
	"github.com/streadway/amqp"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type ExchangeType string

const (
	Direct  ExchangeType = "direct"
	Fanout  ExchangeType = "fanout"
	Topic   ExchangeType = "topic"
	Headers ExchangeType = "headers"
)

func (et ExchangeType) IsValid() error {
	switch et {
	case Direct, Fanout, Topic, Headers:
		return nil
	}
	return errors.New("unknown exchange-type")
}

func (et ExchangeType) String() string {
	switch et {
	case Direct, Fanout, Topic, Headers:
		return string(et)
	default:
		return "direct"
	}
}


type queue struct {
	active                 int32
	pipe                   *jobs.Pipeline
	exchange               string
	exchangeType           ExchangeType
	name, key              string
	consumer               string

	// active consuming channel
	muc sync.Mutex
	cc  *channel

	// queue events
	lsn func(event int, ctx interface{})

	// active operations
	muw sync.RWMutex
	wg  sync.WaitGroup

	// exec handlers
	running    int32
	execPool   chan jobs.Handler
	errHandler jobs.ErrorHandler
}

// newQueue creates new queue wrapper for AMQP.
func newQueue(pipe *jobs.Pipeline, lsn func(event int, ctx interface{})) (*queue, error) {
	if pipe.String("queue", "") == "" {
		return nil, fmt.Errorf("missing `queue` parameter on amqp pipeline")
	}

	exchangeType := ExchangeType(pipe.String("exchange-type", "direct"))

	err := exchangeType.IsValid()
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return &queue{
		exchange: pipe.String("exchange", "amqp.direct"),
		exchangeType: exchangeType,
		name:     pipe.String("queue", ""),
		key:      pipe.String("routing-key", pipe.String("queue", "")),
		consumer: pipe.String("consumer", fmt.Sprintf("rr-jobs:%s-%v", pipe.Name(), os.Getpid())),
		pipe:     pipe,
		lsn:      lsn,
	}, nil
}

// serve consumes queue
func (q *queue) serve(publish, consume *chanPool) {
	atomic.StoreInt32(&q.active, 1)

	for {
		<-consume.waitConnected()
		if atomic.LoadInt32(&q.active) == 0 {
			// stopped
			return
		}

		delivery, cc, err := q.consume(consume)
		if err != nil {
			q.report(err)
			continue
		}

		q.muc.Lock()
		q.cc = cc
		q.muc.Unlock()

		for d := range delivery {
			q.muw.Lock()
			q.wg.Add(1)
			q.muw.Unlock()

			atomic.AddInt32(&q.running, 1)
			h := <-q.execPool

			go func(h jobs.Handler, d amqp.Delivery) {
				err := q.do(publish, h, d)

				atomic.AddInt32(&q.running, ^int32(0))
				q.execPool <- h
				q.wg.Done()
				q.report(err)
			}(h, d)
		}
	}
}

func (q *queue) consume(consume *chanPool) (jobs <-chan amqp.Delivery, cc *channel, err error) {
	// allocate channel for the consuming
	if cc, err = consume.channel(q.name); err != nil {
		return nil, nil, err
	}

	if err := cc.ch.Qos(q.pipe.Integer("prefetch", 4), 0, false); err != nil {
		return nil, nil, consume.closeChan(cc, err)
	}

	delivery, err := cc.ch.Consume(q.name, q.consumer, false, false, false, false, nil)
	if err != nil {
		return nil, nil, consume.closeChan(cc, err)
	}

	// do i like it?
	go func(consume *chanPool) {
		for err := range cc.signal {
			consume.closeChan(cc, err)
			return
		}
	}(consume)

	return delivery, cc, err
}

func (q *queue) do(cp *chanPool, h jobs.Handler, d amqp.Delivery) error {
	id, attempt, j, err := unpack(d)
	if err != nil {
		q.report(err)
		return d.Nack(false, false)
	}
	err = h(id, j)

	if err == nil {
		return d.Ack(false)
	}

	// failed
	q.errHandler(id, j, err)

	if !j.Options.CanRetry(attempt) {
		return d.Nack(false, false)
	}

	// retry as new j (to accommodate attempt number and new delay)
	if err = q.publish(cp, id, attempt+1, j, j.Options.RetryDuration()); err != nil {
		q.report(err)
		return d.Nack(false, true)
	}

	return d.Ack(false)
}

func (q *queue) stop() {
	if atomic.LoadInt32(&q.active) == 0 {
		return
	}

	atomic.StoreInt32(&q.active, 0)

	q.muc.Lock()
	if q.cc != nil {
		// gracefully stopped consuming
		q.report(q.cc.ch.Cancel(q.consumer, true))
	}
	q.muc.Unlock()

	q.muw.Lock()
	q.wg.Wait()
	q.muw.Unlock()
}

// publish message to queue or to delayed queue.
func (q *queue) publish(cp *chanPool, id string, attempt int, j *jobs.Job, delay time.Duration) error {
	c, err := cp.channel(q.name)
	if err != nil {
		return err
	}

	qKey := q.key

	if delay != 0 {
		delayMs := int64(delay.Seconds() * 1000)
		qName := fmt.Sprintf("delayed-%d.%s.%s", delayMs, q.exchange, q.name)
		qKey = qName

		err := q.declare(cp, qName, qName, amqp.Table{
			"x-dead-letter-exchange":    q.exchange,
			"x-dead-letter-routing-key": q.name,
			"x-message-ttl":             delayMs,
			"x-expires":                 delayMs * 2,
		})

		if err != nil {
			return err
		}
	}

	err = c.ch.Publish(
		q.exchange, // exchange
		qKey,      // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/octet-stream",
			Body:         j.Body(),
			DeliveryMode: amqp.Persistent,
			Headers:      pack(id, attempt, j),
		},
	)

	if err != nil {
		return cp.closeChan(c, err)
	}

	confirmed, ok := <-c.confirm
	if ok && confirmed.Ack {
		return nil
	}

	return fmt.Errorf("failed to publish: %v", confirmed.DeliveryTag)
}

// declare queue and binding to it
func (q *queue) declare(cp *chanPool, queue string, key string, args amqp.Table) error {
	c, err := cp.channel(q.name)
	if err != nil {
		return err
	}

	err = c.ch.ExchangeDeclare(q.exchange, q.exchangeType.String(), true, false, false, false, nil)
	if err != nil {
		return cp.closeChan(c, err)
	}

	_, err = c.ch.QueueDeclare(queue, true, false, false, false, args)
	if err != nil {
		return cp.closeChan(c, err)
	}

	err = c.ch.QueueBind(queue, key, q.exchange, false, nil)
	if err != nil {
		return cp.closeChan(c, err)
	}

	// keep channel open
	return err
}

// inspect the queue
func (q *queue) inspect(cp *chanPool) (*amqp.Queue, error) {
	c, err := cp.channel("stat")
	if err != nil {
		return nil, err
	}

	queue, err := c.ch.QueueInspect(q.name)
	if err != nil {
		return nil, cp.closeChan(c, err)
	}

	// keep channel open
	return &queue, err
}

// throw handles service, server and pool events.
func (q *queue) report(err error) {
	if err != nil {
		q.lsn(jobs.EventPipeError, &jobs.PipelineError{Pipeline: q.pipe, Caused: err})
	}
}
