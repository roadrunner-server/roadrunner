package ephemeral

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/spiral/jobs/v2"
	"sync"
)

// Broker run queue using local goroutines.
type Broker struct {
	lsn     func(event int, ctx interface{})
	mu      sync.Mutex
	wait    chan error
	stopped chan interface{}
	queues  map[*jobs.Pipeline]*queue
}

// Listen attaches server event watcher.
func (b *Broker) Listen(lsn func(event int, ctx interface{})) {
	b.lsn = lsn
}

// Init configures broker.
func (b *Broker) Init() (bool, error) {
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

	b.queues[pipe] = newQueue(pipe.Integer("maxThreads", 0))

	return nil
}

// Serve broker pipelines.
func (b *Broker) Serve() error {
	// start consuming
	b.mu.Lock()
	for _, q := range b.queues {
		qq := q
		if qq.execPool != nil {
			go qq.serve()
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

	// stop all consuming
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

	if b.wait != nil {
		if q.execPool != nil {
			go q.serve()
		}
	}

	return nil
}

// Push job into the worker.
func (b *Broker) Push(pipe *jobs.Pipeline, j *jobs.Job) (string, error) {
	if err := b.isServing(); err != nil {
		return "", err
	}

	q := b.queue(pipe)
	if q == nil {
		return "", fmt.Errorf("undefined queue `%s`", pipe.Name())
	}

	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	q.push(id.String(), j, 0, j.Options.DelayDuration())

	return id.String(), nil
}

// Stat must consume statistics about given pipeline or return error.
func (b *Broker) Stat(pipe *jobs.Pipeline) (stat *jobs.Stat, err error) {
	if err := b.isServing(); err != nil {
		return nil, err
	}

	q := b.queue(pipe)
	if q == nil {
		return nil, fmt.Errorf("undefined queue `%s`", pipe.Name())
	}

	return q.stat(), nil
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
