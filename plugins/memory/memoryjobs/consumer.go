package memoryjobs

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	jobState "github.com/spiral/roadrunner/v2/pkg/state/job"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/job"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/utils"
)

const (
	prefetch      string = "prefetch"
	goroutinesMax uint64 = 1000
)

type Config struct {
	Prefetch uint64 `mapstructure:"prefetch"`
}

type consumer struct {
	cfg           *Config
	log           logger.Logger
	eh            events.Handler
	pipeline      atomic.Value
	pq            priorityqueue.Queue
	localPrefetch chan *Item

	// time.sleep goroutines max number
	goroutines uint64

	delayed *int64
	active  *int64

	listeners uint32
	stopCh    chan struct{}
}

func NewJobBroker(configKey string, log logger.Logger, cfg config.Configurer, eh events.Handler, pq priorityqueue.Queue) (*consumer, error) {
	const op = errors.Op("new_ephemeral_pipeline")

	jb := &consumer{
		log:        log,
		pq:         pq,
		eh:         eh,
		goroutines: 0,
		active:     utils.Int64(0),
		delayed:    utils.Int64(0),
		stopCh:     make(chan struct{}),
	}

	err := cfg.UnmarshalKey(configKey, &jb.cfg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if jb.cfg == nil {
		return nil, errors.E(op, errors.Errorf("config not found by provided key: %s", configKey))
	}

	if jb.cfg.Prefetch == 0 {
		jb.cfg.Prefetch = 100_000
	}

	// initialize a local queue
	jb.localPrefetch = make(chan *Item, jb.cfg.Prefetch)

	return jb, nil
}

func FromPipeline(pipeline *pipeline.Pipeline, log logger.Logger, eh events.Handler, pq priorityqueue.Queue) (*consumer, error) {
	return &consumer{
		log:           log,
		pq:            pq,
		eh:            eh,
		localPrefetch: make(chan *Item, pipeline.Int(prefetch, 100_000)),
		goroutines:    0,
		active:        utils.Int64(0),
		delayed:       utils.Int64(0),
		stopCh:        make(chan struct{}),
	}, nil
}

func (c *consumer) Push(ctx context.Context, jb *job.Job) error {
	const op = errors.Op("ephemeral_push")

	// check if the pipeline registered
	_, ok := c.pipeline.Load().(*pipeline.Pipeline)
	if !ok {
		return errors.E(op, errors.Errorf("no such pipeline: %s", jb.Options.Pipeline))
	}

	err := c.handleItem(ctx, fromJob(jb))
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (c *consumer) State(_ context.Context) (*jobState.State, error) {
	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	return &jobState.State{
		Pipeline: pipe.Name(),
		Driver:   pipe.Driver(),
		Queue:    pipe.Name(),
		Active:   atomic.LoadInt64(c.active),
		Delayed:  atomic.LoadInt64(c.delayed),
		Ready:    ready(atomic.LoadUint32(&c.listeners)),
	}, nil
}

func (c *consumer) Register(_ context.Context, pipeline *pipeline.Pipeline) error {
	c.pipeline.Store(pipeline)
	return nil
}

func (c *consumer) Pause(_ context.Context, p string) {
	start := time.Now()
	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		c.log.Error("no such pipeline", "requested pause on: ", p)
	}

	l := atomic.LoadUint32(&c.listeners)
	// no active listeners
	if l == 0 {
		c.log.Warn("no active listeners, nothing to pause")
		return
	}

	atomic.AddUint32(&c.listeners, ^uint32(0))

	// stop the consumer
	c.stopCh <- struct{}{}

	c.eh.Push(events.JobEvent{
		Event:    events.EventPipePaused,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    start,
		Elapsed:  time.Since(start),
	})
}

func (c *consumer) Resume(_ context.Context, p string) {
	start := time.Now()
	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		c.log.Error("no such pipeline", "requested resume on: ", p)
	}

	l := atomic.LoadUint32(&c.listeners)
	// listener already active
	if l == 1 {
		c.log.Warn("listener already in the active state")
		return
	}

	// resume the consumer on the same channel
	c.consume()

	atomic.StoreUint32(&c.listeners, 1)
	c.eh.Push(events.JobEvent{
		Event:    events.EventPipeActive,
		Pipeline: pipe.Name(),
		Driver:   pipe.Driver(),
		Start:    start,
		Elapsed:  time.Since(start),
	})
}

// Run is no-op for the ephemeral
func (c *consumer) Run(_ context.Context, pipe *pipeline.Pipeline) error {
	c.eh.Push(events.JobEvent{
		Event:    events.EventPipeActive,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    time.Now(),
	})
	return nil
}

func (c *consumer) Stop(_ context.Context) error {
	start := time.Now()
	pipe := c.pipeline.Load().(*pipeline.Pipeline)

	select {
	case c.stopCh <- struct{}{}:
	default:
		break
	}

	for i := 0; i < len(c.localPrefetch); i++ {
		// drain all jobs from the channel
		<-c.localPrefetch
	}

	c.localPrefetch = nil

	c.eh.Push(events.JobEvent{
		Event:    events.EventPipeStopped,
		Pipeline: pipe.Name(),
		Driver:   pipe.Driver(),
		Start:    start,
		Elapsed:  time.Since(start),
	})

	return nil
}

func (c *consumer) handleItem(ctx context.Context, msg *Item) error {
	const op = errors.Op("ephemeral_handle_request")
	// handle timeouts
	// theoretically, some bad user may send millions requests with a delay and produce a billion (for example)
	// goroutines here. We should limit goroutines here.
	if msg.Options.Delay > 0 {
		// if we have 1000 goroutines waiting on the delay - reject 1001
		if atomic.LoadUint64(&c.goroutines) >= goroutinesMax {
			return errors.E(op, errors.Str("max concurrency number reached"))
		}

		go func(jj *Item) {
			atomic.AddUint64(&c.goroutines, 1)
			atomic.AddInt64(c.delayed, 1)

			time.Sleep(jj.Options.DelayDuration())

			select {
			case c.localPrefetch <- jj:
				atomic.AddUint64(&c.goroutines, ^uint64(0))
			default:
				c.log.Warn("can't push job", "error", "local queue closed or full")
			}
		}(msg)

		return nil
	}

	// increase number of the active jobs
	atomic.AddInt64(c.active, 1)

	// insert to the local, limited pipeline
	select {
	case c.localPrefetch <- msg:
		return nil
	case <-ctx.Done():
		return errors.E(op, errors.Errorf("local pipeline is full, consider to increase prefetch number, current limit: %d, context error: %v", c.cfg.Prefetch, ctx.Err()))
	}
}

func (c *consumer) consume() {
	go func() {
		// redirect
		for {
			select {
			case item, ok := <-c.localPrefetch:
				if !ok {
					c.log.Warn("ephemeral local prefetch queue closed")
					return
				}

				// set requeue channel
				item.Options.requeueFn = c.handleItem
				item.Options.active = c.active
				item.Options.delayed = c.delayed

				c.pq.Insert(item)
			case <-c.stopCh:
				return
			}
		}
	}()
}

func ready(r uint32) bool {
	return r > 0
}
