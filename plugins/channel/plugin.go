package channel

import (
	"sync"
)

const (
	PluginName string = "hub"
)

type Plugin struct {
	sync.Mutex
	send    chan interface{}
	receive chan interface{}
}

func (p *Plugin) Init() error {
	p.Lock()
	defer p.Unlock()
	p.send = make(chan interface{})
	p.receive = make(chan interface{})
	return nil
}

func (p *Plugin) Serve() chan error {
	return make(chan error)
}

func (p *Plugin) Stop() error {
	close(p.receive)
	return nil
}

func (p *Plugin) SendCh() chan interface{} {
	p.Lock()
	defer p.Unlock()
	// bi-directional queue
	return p.send
}

func (p *Plugin) ReceiveCh() chan interface{} {
	p.Lock()
	defer p.Unlock()
	// bi-directional queue
	return p.receive
}

func (p *Plugin) Available() {}

func (p *Plugin) Name() string {
	return PluginName
}
