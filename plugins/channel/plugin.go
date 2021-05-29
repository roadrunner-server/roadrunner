package channel

import (
	"sync"
)

const (
	PluginName string = "hub"
)

type Plugin struct {
	sync.Mutex
	fromCh chan interface{}
	toCh   chan interface{}
}

func (p *Plugin) Init() error {
	p.Lock()
	defer p.Unlock()

	p.fromCh = make(chan interface{})
	p.toCh = make(chan interface{})
	return nil
}

func (p *Plugin) Serve() chan error {
	return make(chan error)
}

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) FromWorker() chan interface{} {
	p.Lock()
	defer p.Unlock()
	// one-directional queue
	return p.fromCh
}

func (p *Plugin) ToWorker() chan interface{} {
	p.Lock()
	defer p.Unlock()
	// one-directional queue
	return p.toCh
}

func (p *Plugin) Name() string {
	return PluginName
}
