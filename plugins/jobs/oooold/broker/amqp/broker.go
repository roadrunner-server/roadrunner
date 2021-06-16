package amqp

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/spiral/jobs/v2"
	"sync"
	"sync/atomic"
)

// Broker represents AMQP broker.
type Broker struct {
	cfg     *Config
	lsn     func(event int, ctx interface{})
	publish *chanPool
	consume *chanPool
	mu      sync.Mutex
	wait    chan error
	stopped chan interface{}
	queues  map[*jobs.Pipeline]*queue
}

// Listen attaches server event watcher.
func (b *Broker) Listen(lsn func(event int, ctx interface{})) {
	b.lsn = lsn
}

// Init configures AMQP job broker (always 2 connections).
func (b *Broker) Init(cfg *Config) (ok bool, err error) {
	b.cfg = cfg
	b.queues = make(map[*jobs.Pipeline]*queue)

	return true, nil
}

// Register broker pipeline.
func (b *Broker) Register(pipe *jobs.Pipeline) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.queues[pipe]; ok {
		return fmt.Errorf("queue `%s` has already been registered", pipe.Name())
	}

	q, err := newQueue(pipe, b.throw)
	if err != nil {
		return err
	}

	b.queues[pipe] = q

	return nil
}

// Serve broker pipelines.
func (b *Broker) Serve() (err error) {
	b.mu.Lock()

	if b.publish, err = newConn(b.cfg.Addr, b.cfg.TimeoutDuration()); err != nil {
		b.mu.Unlock()
		return err
	}
	defer b.publish.Close()

	if b.consume, err = newConn(b.cfg.Addr, b.cfg.TimeoutDuration()); err != nil {
		b.mu.Unlock()
		return err
	}
	defer b.consume.Close()

	for _, q := range b.queues {
		err := q.declare(b.publish, q.name, q.key, nil)
		if err != nil {
			b.mu.Unlock()
			return err
		}
	}

	for _, q := range b.queues {
		qq := q
		if qq.execPool != nil {
			go qq.serve(b.publish, b.consume)
		}
	}

	b.wait = make(chan error)
	b.stopped = make(chan interface{})
	defer close(b.stopped)

	b.mu.Unlock()

	b.throw(jobs.EventBrokerReady, b)

	return <-b.wait
}

// Stop all pipelines.
func (b *Broker) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.wait == nil {
		return
	}

	for _, q := range b.queues {
		q.stop()
	}

	close(b.wait)
	<-b.stopped
}

// Consume configures pipeline to be consumed. With execPool to nil to disable consuming. Method can be called before
// the service is started!
func (b *Broker) Consume(pipe *jobs.Pipeline, execPool chan jobs.Handler, errHandler jobs.ErrorHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	q, ok := b.queues[pipe]
	if !ok {
		return fmt.Errorf("undefined queue `%s`", pipe.Name())
	}

	q.stop()

	q.execPool = execPool
	q.errHandler = errHandler

	if b.publish != nil && q.execPool != nil {
		if q.execPool != nil {
			go q.serve(b.publish, b.consume)
		}
	}

	return nil
}

// Push job into the worker.
func (b *Broker) Push(pipe *jobs.Pipeline, j *jobs.Job) (string, error) {
	if err := b.isServing(); err != nil {
		return "", err
	}

	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	q := b.queue(pipe)
	if q == nil {
		return "", fmt.Errorf("undefined queue `%s`", pipe.Name())
	}

	if err := q.publish(b.publish, id.String(), 0, j, j.Options.DelayDuration()); err != nil {
		return "", err
	}

	return id.String(), nil
}

// Stat must fetch statistics about given pipeline or return error.
func (b *Broker) Stat(pipe *jobs.Pipeline) (stat *jobs.Stat, err error) {
	if err := b.isServing(); err != nil {
		return nil, err
	}

	q := b.queue(pipe)
	if q == nil {
		return nil, fmt.Errorf("undefined queue `%s`", pipe.Name())
	}

	queue, err := q.inspect(b.publish)
	if err != nil {
		return nil, err
	}

	// this the closest approximation we can get for now
	return &jobs.Stat{
		InternalName: queue.Name,
		Queue:        int64(queue.Messages),
		Active:       int64(atomic.LoadInt32(&q.running)),
	}, nil
}

// check if broker is serving
func (b *Broker) isServing() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.wait == nil {
		return fmt.Errorf("broker is not running")
	}

	return nil
}

// queue returns queue associated with the pipeline.
func (b *Broker) queue(pipe *jobs.Pipeline) *queue {
	b.mu.Lock()
	defer b.mu.Unlock()

	q, ok := b.queues[pipe]
	if !ok {
		return nil
	}

	return q
}

// throw handles service, server and pool events.
func (b *Broker) throw(event int, ctx interface{}) {
	if b.lsn != nil {
		b.lsn(event, ctx)
	}
}
