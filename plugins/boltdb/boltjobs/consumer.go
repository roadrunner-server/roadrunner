package boltjobs

import (
	"bytes"
	"context"
	"encoding/gob"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	jobState "github.com/spiral/roadrunner/v2/pkg/state/job"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/job"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/utils"
	bolt "go.etcd.io/bbolt"
)

const (
	PluginName = "boltdb"

	PushBucket    = "push"
	InQueueBucket = "processing"
	DoneBucket    = "done"
)

type consumer struct {
	file        string
	permissions int
	priority    int
	prefetch    int

	db *bolt.DB

	log       logger.Logger
	eh        events.Handler
	pq        priorityqueue.Queue
	listeners uint32
	pipeline  atomic.Value

	stopCh chan struct{}
}

func NewBoltDBJobs(configKey string, log logger.Logger, cfg config.Configurer, e events.Handler, pq priorityqueue.Queue) (*consumer, error) {
	const op = errors.Op("init_boltdb_jobs")

	if !cfg.Has(configKey) {
		return nil, errors.E(op, errors.Errorf("no configuration by provided key: %s", configKey))
	}

	// if no global section
	if !cfg.Has(PluginName) {
		return nil, errors.E(op, errors.Str("no global boltdb configuration"))
	}

	conf := &GlobalCfg{}

	err := cfg.UnmarshalKey(PluginName, conf)
	if err != nil {
		return nil, errors.E(op, err)
	}

	localCfg := &Config{}
	err = cfg.UnmarshalKey(configKey, localCfg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	localCfg.InitDefaults()
	conf.InitDefaults()

	db, err := bolt.Open(localCfg.File, os.FileMode(conf.Permissions), &bolt.Options{
		Timeout:        time.Second * 20,
		NoGrowSync:     false,
		NoFreelistSync: false,
		ReadOnly:       false,
		NoSync:         false,
	})

	if err != nil {
		return nil, errors.E(op, err)
	}

	// create bucket if it does not exist
	// tx.Commit invokes via the db.Update
	err = db.Update(func(tx *bolt.Tx) error {
		const upOp = errors.Op("boltdb_plugin_update")
		_, err = tx.CreateBucketIfNotExists(utils.AsBytes(PushBucket))
		if err != nil {
			return errors.E(op, upOp)
		}
		_, err = tx.CreateBucketIfNotExists(utils.AsBytes(InQueueBucket))
		if err != nil {
			return errors.E(op, upOp)
		}
		_, err = tx.CreateBucketIfNotExists(utils.AsBytes(DoneBucket))
		if err != nil {
			return errors.E(op, upOp)
		}
		return nil
	})
	if err != nil {
		return nil, errors.E(op, err)
	}

	return &consumer{
		permissions: conf.Permissions,
		file:        localCfg.File,
		priority:    localCfg.Priority,
		prefetch:    localCfg.Prefetch,

		db:     db,
		log:    log,
		eh:     e,
		pq:     pq,
		stopCh: make(chan struct{}, 1),
	}, nil
}

func FromPipeline(pipeline *pipeline.Pipeline, log logger.Logger, cfg config.Configurer, e events.Handler, pq priorityqueue.Queue) (*consumer, error) {
	const op = errors.Op("init_boltdb_jobs")

	// if no global section
	if !cfg.Has(PluginName) {
		return nil, errors.E(op, errors.Str("no global boltdb configuration"))
	}

	conf := &GlobalCfg{}
	err := cfg.UnmarshalKey(PluginName, conf)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// add default values
	conf.InitDefaults()

	db, err := bolt.Open(pipeline.String(file, "rr.db"), os.FileMode(conf.Permissions), &bolt.Options{
		Timeout:        time.Second * 20,
		NoGrowSync:     false,
		NoFreelistSync: false,
		ReadOnly:       false,
		NoSync:         false,
	})

	if err != nil {
		return nil, errors.E(op, err)
	}

	// create bucket if it does not exist
	// tx.Commit invokes via the db.Update
	err = db.Update(func(tx *bolt.Tx) error {
		const upOp = errors.Op("boltdb_plugin_update")
		_, err = tx.CreateBucketIfNotExists(utils.AsBytes(PushBucket))
		if err != nil {
			return errors.E(op, upOp)
		}
		_, err = tx.CreateBucketIfNotExists(utils.AsBytes(InQueueBucket))
		if err != nil {
			return errors.E(op, upOp)
		}
		_, err = tx.CreateBucketIfNotExists(utils.AsBytes(DoneBucket))
		if err != nil {
			return errors.E(op, upOp)
		}
		return nil
	})

	if err != nil {
		return nil, errors.E(op, err)
	}

	return &consumer{
		file:        pipeline.String(file, "rr.db"),
		priority:    pipeline.Int(priority, 10),
		prefetch:    pipeline.Int(prefetch, 100),
		permissions: conf.Permissions,

		db:     db,
		log:    log,
		eh:     e,
		pq:     pq,
		stopCh: make(chan struct{}, 1),
	}, nil
}

func (c *consumer) Push(ctx context.Context, job *job.Job) error {
	const op = errors.Op("boltdb_jobs_push")
	err := c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(utils.AsBytes(PushBucket))
		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		err := enc.Encode(job)
		if err != nil {
			return err
		}

		return b.Put(utils.AsBytes(uuid.NewString()), buf.Bytes())
	})

	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (c *consumer) Register(_ context.Context, pipeline *pipeline.Pipeline) error {
	c.pipeline.Store(pipeline)
	return nil
}

func (c *consumer) Run(_ context.Context, p *pipeline.Pipeline) error {
	const op = errors.Op("boltdb_run")

	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p.Name() {
		return errors.E(op, errors.Errorf("no such pipeline registered: %s", pipe.Name()))
	}
	return nil
}

func (c *consumer) Stop(ctx context.Context) error {
	return nil
}

func (c *consumer) Pause(ctx context.Context, p string) {
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

	c.stopCh <- struct{}{}

	atomic.AddUint32(&c.listeners, ^uint32(0))

	c.eh.Push(events.JobEvent{
		Event:    events.EventPipePaused,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    time.Now(),
	})
}

func (c *consumer) Resume(ctx context.Context, p string) {
	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		c.log.Error("no such pipeline", "requested resume on: ", p)
	}

	l := atomic.LoadUint32(&c.listeners)
	// no active listeners
	if l == 1 {
		c.log.Warn("amqp listener already in the active state")
		return
	}

	// run listener
	go c.listener()

	// increase number of listeners
	atomic.AddUint32(&c.listeners, 1)

	c.eh.Push(events.JobEvent{
		Event:    events.EventPipeActive,
		Driver:   pipe.Driver(),
		Pipeline: pipe.Name(),
		Start:    time.Now(),
	})
}

func (c *consumer) State(ctx context.Context) (*jobState.State, error) {
	return nil, nil
}
