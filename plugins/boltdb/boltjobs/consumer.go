package boltjobs

import (
	"context"
	"os"
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
	bolt "go.etcd.io/bbolt"
)

const (
	PluginName = "boltdb"
)

type consumer struct {
	// bbolt configuration
	file        string
	permissions int
	bucket      string
	db          *bolt.DB

	log  logger.Logger
	eh   events.Handler
	pq   priorityqueue.Queue
	pipe atomic.Value
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

	conf := &Config{}

	err := cfg.UnmarshalKey(configKey, conf)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// add default values
	conf.InitDefaults()
	c := &consumer{
		file:        conf.File,
		permissions: conf.Permissions,
		bucket:      conf.bucket,

		log: log,
		eh:  e,
		pq:  pq,
	}

	db, err := bolt.Open(c.file, os.FileMode(c.permissions), &bolt.Options{
		Timeout:        time.Second * 20,
		NoGrowSync:     false,
		NoFreelistSync: false,
		ReadOnly:       false,
		NoSync:         false,
	})

	if err != nil {
		return nil, errors.E(op, err)
	}

	c.db = db

	// create bucket if it does not exist
	// tx.Commit invokes via the db.Update
	err = db.Update(func(tx *bolt.Tx) error {
		const upOp = errors.Op("boltdb_plugin_update")
		_, err = tx.CreateBucketIfNotExists(utils.AsBytes(c.bucket))
		if err != nil {
			return errors.E(op, upOp)
		}
		return nil
	})

	return c, nil
}

func FromPipeline(pipeline *pipeline.Pipeline, log logger.Logger, cfg config.Configurer, e events.Handler, pq priorityqueue.Queue) (*consumer, error) {
	return &consumer{}, nil
}

func (c *consumer) Push(ctx context.Context, job *job.Job) error {
	panic("implement me")
}

func (c *consumer) Register(_ context.Context, pipeline *pipeline.Pipeline) error {
	c.pipe.Store(pipeline)
	return nil
}

func (c *consumer) Run(_ context.Context, pipeline *pipeline.Pipeline) error {
	panic("implement me")
}

func (c *consumer) Stop(ctx context.Context) error {
	panic("implement me")
}

func (c *consumer) Pause(ctx context.Context, pipeline string) {
	panic("implement me")
}

func (c *consumer) Resume(ctx context.Context, pipeline string) {
	panic("implement me")
}

func (c *consumer) State(ctx context.Context) (*jobState.State, error) {
	panic("implement me")
}
