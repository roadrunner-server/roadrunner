package oooold

import (
	"fmt"
	"github.com/gofrs/uuid"
	"sync"
	"sync/atomic"
	"time"
)

// testBroker run testQueue using local goroutines.
type testBroker struct {
	lsn     func(event int, ctx interface{})
	mu      sync.Mutex
	wait    chan error
	stopped chan interface{}
	queues  map[*Pipeline]*testQueue
}

// Listen attaches server event watcher.
func (b *testBroker) Listen(lsn func(event int, ctx interface{})) {
	b.lsn = lsn
}

// Init configures broker.
func (b *testBroker) Init() (bool, error) {
	b.queues = make(map[*Pipeline]*testQueue)

	return true, nil
}

// Register broker pipeline.
func (b *testBroker) Register(pipe *Pipeline) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.queues[pipe]; ok {
		return fmt.Errorf("testQueue `%s` has already been registered", pipe.Name())
	}

	b.queues[pipe] = newQueue()

	return nil
}

// Serve broker pipelines.
func (b *testBroker) Serve() error {
	// start pipelines
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

	b.throw(EventBrokerReady, b)

	return <-b.wait
}

// Stop all pipelines.
func (b *testBroker) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.wait == nil {
		return
	}

	// stop all pipelines
	for _, q := range b.queues {
		q.stop()
	}

	close(b.wait)
	<-b.stopped
}

// Consume configures pipeline to be consumed. With execPool to nil to disable pipelines. Method can be called before
// the service is started!
func (b *testBroker) Consume(pipe *Pipeline, execPool chan Handler, errHandler ErrorHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	q, ok := b.queues[pipe]
	if !ok {
		return fmt.Errorf("undefined testQueue `%s`", pipe.Name())
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
func (b *testBroker) Push(pipe *Pipeline, j *Job) (string, error) {
	if err := b.isServing(); err != nil {
		return "", err
	}

	q := b.queue(pipe)
	if q == nil {
		return "", fmt.Errorf("undefined testQueue `%s`", pipe.Name())
	}

	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	q.push(id.String(), j, 0, j.Options.DelayDuration())

	return id.String(), nil
}

// Stat must consume statistics about given pipeline or return error.
func (b *testBroker) Stat(pipe *Pipeline) (stat *Stat, err error) {
	if err := b.isServing(); err != nil {
		return nil, err
	}

	q := b.queue(pipe)
	if q == nil {
		return nil, fmt.Errorf("undefined testQueue `%s`", pipe.Name())
	}

	return q.stat(), nil
}

// check if broker is serving
func (b *testBroker) isServing() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.wait == nil {
		return fmt.Errorf("broker is not running")
	}

	return nil
}

// testQueue returns testQueue associated with the pipeline.
func (b *testBroker) queue(pipe *Pipeline) *testQueue {
	b.mu.Lock()
	defer b.mu.Unlock()

	q, ok := b.queues[pipe]
	if !ok {
		return nil
	}

	return q
}

// throw handles service, server and pool events.
func (b *testBroker) throw(event int, ctx interface{}) {
	if b.lsn != nil {
		b.lsn(event, ctx)
	}
}

type testQueue struct {
	active int32
	st     *Stat

	// job pipeline
	jobs chan *entry

	// pipelines operations
	muw sync.Mutex
	wg  sync.WaitGroup

	// stop channel
	wait chan interface{}

	// exec handlers
	execPool   chan Handler
	errHandler ErrorHandler
}

type entry struct {
	id      string
	job     *Job
	attempt int
}

// create new testQueue
func newQueue() *testQueue {
	return &testQueue{st: &Stat{}, jobs: make(chan *entry)}
}

// todo NOT USED
// associate testQueue with new do pool
//func (q *testQueue) configure(execPool chan Handler, err ErrorHandler) error {
//	q.execPool = execPool
//	q.errHandler = err
//
//	return nil
//}

// serve consumers
func (q *testQueue) serve() {
	q.wait = make(chan interface{})
	atomic.StoreInt32(&q.active, 1)

	for {
		e := q.consume()
		if e == nil {
			return
		}

		atomic.AddInt64(&q.st.Active, 1)
		h := <-q.execPool
		go func(e *entry) {
			q.do(h, e)
			atomic.AddInt64(&q.st.Active, ^int64(0))
			q.execPool <- h
			q.wg.Done()
		}(e)
	}
}

// allocate one job entry
func (q *testQueue) consume() *entry {
	q.muw.Lock()
	defer q.muw.Unlock()

	select {
	case <-q.wait:
		return nil
	case e := <-q.jobs:
		q.wg.Add(1)

		return e
	}
}

// do singe job
func (q *testQueue) do(h Handler, e *entry) {
	err := h(e.id, e.job)

	if err == nil {
		atomic.AddInt64(&q.st.Queue, ^int64(0))
		return
	}

	q.errHandler(e.id, e.job, err)

	if !e.job.Options.CanRetry(e.attempt) {
		atomic.AddInt64(&q.st.Queue, ^int64(0))
		return
	}

	q.push(e.id, e.job, e.attempt+1, e.job.Options.RetryDuration())
}

// stop the testQueue pipelines
func (q *testQueue) stop() {
	if atomic.LoadInt32(&q.active) == 0 {
		return
	}

	atomic.StoreInt32(&q.active, 0)

	close(q.wait)
	q.muw.Lock()
	q.wg.Wait()
	q.muw.Unlock()
}

// add job to the testQueue
func (q *testQueue) push(id string, j *Job, attempt int, delay time.Duration) {
	if delay == 0 {
		atomic.AddInt64(&q.st.Queue, 1)
		go func() {
			q.jobs <- &entry{id: id, job: j, attempt: attempt}
		}()

		return
	}

	atomic.AddInt64(&q.st.Delayed, 1)
	go func() {
		time.Sleep(delay)
		atomic.AddInt64(&q.st.Delayed, ^int64(0))
		atomic.AddInt64(&q.st.Queue, 1)

		q.jobs <- &entry{id: id, job: j, attempt: attempt}
	}()
}

func (q *testQueue) stat() *Stat {
	return &Stat{
		InternalName: ":memory:",
		Queue:        atomic.LoadInt64(&q.st.Queue),
		Active:       atomic.LoadInt64(&q.st.Active),
		Delayed:      atomic.LoadInt64(&q.st.Delayed),
	}
}
