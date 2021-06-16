package ephemeral

import (
	"github.com/spiral/jobs/v2"
	"sync"
	"sync/atomic"
	"time"
)

type queue struct {
	on    int32
	state *jobs.Stat

	// job pipeline
	concurPool chan interface{}
	jobs       chan *entry

	// on operations
	muw sync.Mutex
	wg  sync.WaitGroup

	// stop channel
	wait chan interface{}

	// exec handlers
	execPool   chan jobs.Handler
	errHandler jobs.ErrorHandler
}

type entry struct {
	id      string
	job     *jobs.Job
	attempt int
}

// create new queue
func newQueue(maxConcur int) *queue {
	q := &queue{state: &jobs.Stat{}, jobs: make(chan *entry)}

	if maxConcur != 0 {
		q.concurPool = make(chan interface{}, maxConcur)
		for i := 0; i < maxConcur; i++ {
			q.concurPool <- nil
		}
	}

	return q
}

// serve consumers
func (q *queue) serve() {
	q.wait = make(chan interface{})
	atomic.StoreInt32(&q.on, 1)

	for {
		e := q.consume()
		if e == nil {
			q.wg.Wait()
			return
		}

		if q.concurPool != nil {
			<-q.concurPool
		}

		atomic.AddInt64(&q.state.Active, 1)
		h := <-q.execPool

		go func(h jobs.Handler, e *entry) {
			defer q.wg.Done()

			q.do(h, e)
			atomic.AddInt64(&q.state.Active, ^int64(0))

			q.execPool <- h

			if q.concurPool != nil {
				q.concurPool <- nil
			}
		}(h, e)
	}
}

// allocate one job entry
func (q *queue) consume() *entry {
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
func (q *queue) do(h jobs.Handler, e *entry) {
	err := h(e.id, e.job)

	if err == nil {
		atomic.AddInt64(&q.state.Queue, ^int64(0))
		return
	}

	q.errHandler(e.id, e.job, err)

	if !e.job.Options.CanRetry(e.attempt) {
		atomic.AddInt64(&q.state.Queue, ^int64(0))
		return
	}

	q.push(e.id, e.job, e.attempt+1, e.job.Options.RetryDuration())
}

// stop the queue consuming
func (q *queue) stop() {
	if atomic.LoadInt32(&q.on) == 0 {
		return
	}

	close(q.wait)

	q.muw.Lock()
	q.wg.Wait()
	q.muw.Unlock()

	atomic.StoreInt32(&q.on, 0)
}

// add job to the queue
func (q *queue) push(id string, j *jobs.Job, attempt int, delay time.Duration) {
	if delay == 0 {
		atomic.AddInt64(&q.state.Queue, 1)
		go func() {
			q.jobs <- &entry{id: id, job: j, attempt: attempt}
		}()

		return
	}

	atomic.AddInt64(&q.state.Delayed, 1)
	go func() {
		time.Sleep(delay)
		atomic.AddInt64(&q.state.Delayed, ^int64(0))
		atomic.AddInt64(&q.state.Queue, 1)

		q.jobs <- &entry{id: id, job: j, attempt: attempt}
	}()
}

func (q *queue) stat() *jobs.Stat {
	return &jobs.Stat{
		InternalName: ":memory:",
		Queue:        atomic.LoadInt64(&q.state.Queue),
		Active:       atomic.LoadInt64(&q.state.Active),
		Delayed:      atomic.LoadInt64(&q.state.Delayed),
	}
}
