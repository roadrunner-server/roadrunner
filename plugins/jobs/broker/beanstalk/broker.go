package beanstalk

import (
	"fmt"
	"github.com/spiral/jobs/v2"
	"sync"
)

// Broker run consume using Broker service.
type Broker struct {
	cfg     *Config
	lsn     func(event int, ctx interface{})
	mu      sync.Mutex
	wait    chan error
	stopped chan interface{}
	conn    *conn
	tubes   map[*jobs.Pipeline]*tube
}

// Listen attaches server event watcher.
func (b *Broker) Listen(lsn func(event int, ctx interface{})) {
	b.lsn = lsn
}

// Init configures broker.
func (b *Broker) Init(cfg *Config) (bool, error) {
	b.cfg = cfg
	b.tubes = make(map[*jobs.Pipeline]*tube)

	return true, nil
}

// Register broker pipeline.
func (b *Broker) Register(pipe *jobs.Pipeline) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.tubes[pipe]; ok {
		return fmt.Errorf("tube `%s` has already been registered", pipe.Name())
	}

	t, err := newTube(pipe, b.throw)
	if err != nil {
		return err
	}

	b.tubes[pipe] = t

	return nil
}

// Serve broker pipelines.
func (b *Broker) Serve() (err error) {
	b.mu.Lock()

	if b.conn, err = b.cfg.newConn(); err != nil {
		return err
	}
	defer b.conn.Close()

	for _, t := range b.tubes {
		tt := t
		if tt.execPool != nil {
			go tt.serve(b.cfg)
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

	for _, t := range b.tubes {
		t.stop()
	}

	close(b.wait)
	<-b.stopped
}

// Consume configures pipeline to be consumed. With execPool to nil to reset consuming. Method can be called before
// the service is started!
func (b *Broker) Consume(pipe *jobs.Pipeline, execPool chan jobs.Handler, errHandler jobs.ErrorHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	t, ok := b.tubes[pipe]
	if !ok {
		return fmt.Errorf("undefined tube `%s`", pipe.Name())
	}

	t.stop()

	t.execPool = execPool
	t.errHandler = errHandler

	if b.conn != nil {
		tt := t
		if tt.execPool != nil {
			go tt.serve(connFactory(b.cfg))
		}
	}

	return nil
}

// Push data into the worker.
func (b *Broker) Push(pipe *jobs.Pipeline, j *jobs.Job) (string, error) {
	if err := b.isServing(); err != nil {
		return "", err
	}

	t := b.tube(pipe)
	if t == nil {
		return "", fmt.Errorf("undefined tube `%s`", pipe.Name())
	}

	data, err := pack(j)
	if err != nil {
		return "", err
	}

	return t.put(b.conn, 0, data, j.Options.DelayDuration(), j.Options.TimeoutDuration())
}

// Stat must fetch statistics about given pipeline or return error.
func (b *Broker) Stat(pipe *jobs.Pipeline) (stat *jobs.Stat, err error) {
	if err := b.isServing(); err != nil {
		return nil, err
	}

	t := b.tube(pipe)
	if t == nil {
		return nil, fmt.Errorf("undefined tube `%s`", pipe.Name())
	}

	return t.stat(b.conn)
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
func (b *Broker) tube(pipe *jobs.Pipeline) *tube {
	b.mu.Lock()
	defer b.mu.Unlock()

	t, ok := b.tubes[pipe]
	if !ok {
		return nil
	}

	return t
}

// throw handles service, server and pool events.
func (b *Broker) throw(event int, ctx interface{}) {
	if b.lsn != nil {
		b.lsn(event, ctx)
	}
}
