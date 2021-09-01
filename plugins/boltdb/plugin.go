package boltdb

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/common/jobs"
	"github.com/spiral/roadrunner/v2/common/kv"
	"github.com/spiral/roadrunner/v2/pkg/events"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/boltdb/boltjobs"
	"github.com/spiral/roadrunner/v2/plugins/boltdb/boltkv"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	PluginName string = "boltdb"
)

// Plugin BoltDB K/V storage.
type Plugin struct {
	cfg config.Configurer
	// logger
	log logger.Logger
}

func (p *Plugin) Init(log logger.Logger, cfg config.Configurer) error {
	p.log = log
	p.cfg = cfg
	return nil
}

// Serve is noop here
func (p *Plugin) Serve() chan error {
	return make(chan error, 1)
}

func (p *Plugin) Stop() error {
	return nil
}

// Name returns plugin name
func (p *Plugin) Name() string {
	return PluginName
}

// Available interface implementation
func (p *Plugin) Available() {}

func (p *Plugin) KVConstruct(key string) (kv.Storage, error) {
	const op = errors.Op("boltdb_plugin_provide")
	st, err := boltkv.NewBoltDBDriver(p.log, key, p.cfg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return st, nil
}

// JOBS bbolt implementation

func (p *Plugin) JobsConstruct(configKey string, e events.Handler, queue priorityqueue.Queue) (jobs.Consumer, error) {
	return boltjobs.NewBoltDBJobs(configKey, p.log, p.cfg, e, queue)
}

func (p *Plugin) FromPipeline(pipe *pipeline.Pipeline, e events.Handler, queue priorityqueue.Queue) (jobs.Consumer, error) {
	return boltjobs.FromPipeline(pipe, p.log, p.cfg, e, queue)
}
