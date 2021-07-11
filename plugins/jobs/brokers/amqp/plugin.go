package amqp

import (
	"sync/atomic"

	"github.com/spiral/roadrunner/v2/common/jobs"
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	pluginName string = "amqp"
)

type Plugin struct {
	log logger.Logger
	cfg config.Configurer

	numConsumers uint32
	stopCh       chan struct{}
}

func (p *Plugin) Init(log logger.Logger, cfg config.Configurer) error {
	p.log = log
	p.cfg = cfg
	p.stopCh = make(chan struct{})
	return nil
}

func (p *Plugin) Serve() chan error {
	return make(chan error)
}

func (p *Plugin) Stop() error {
	// send stop to the all consumers delivery
	for i := uint32(0); i < atomic.LoadUint32(&p.numConsumers); i++ {
		p.stopCh <- struct{}{}
	}
	return nil
}

func (p *Plugin) Name() string {
	return pluginName
}

func (p *Plugin) Available() {}

func (p *Plugin) JobsConstruct(configKey string, pq priorityqueue.Queue) (jobs.Consumer, error) {
	atomic.AddUint32(&p.numConsumers, 1)
	return NewAMQPConsumer(configKey, p.log, p.cfg, p.stopCh, pq)
}
