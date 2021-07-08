package amqp

import (
	"github.com/spiral/roadrunner/v2/common/jobs"
	"github.com/spiral/roadrunner/v2/pkg/priorityqueue"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	name string = "amqp"
)

type Plugin struct {
	log logger.Logger
	cfg config.Configurer
}

func (p *Plugin) Init(log logger.Logger, cfg config.Configurer) error {
	p.log = log
	p.cfg = cfg
	return nil
}

func (p *Plugin) Name() string {
	return name
}

func (p *Plugin) JobsConstruct(configKey string, pq priorityqueue.Queue) (jobs.Consumer, error) {
	return NewAMQPConsumer(configKey, p.log, p.cfg, pq)
}