package pq_plugin //nolint:stylecheck

import (
	priorityqueue "github.com/spiral/roadrunner/v2/pkg/priority_queue"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	PluginName string = "internal_pq"
)

type Plugin struct {
	log logger.Logger
	pq  priorityqueue.Queue
}

func (p *Plugin) Init(log logger.Logger) error {
	p.log = log
	p.pq = priorityqueue.NewPriorityQueue()
	return nil
}

func (p *Plugin) Push(item interface{}) {
	p.pq.Push(item)
	// no-op
}

func (p *Plugin) Pop() interface{} {
	return p.pq.Pop()
}

func (p *Plugin) Name() string {
	return PluginName
}
