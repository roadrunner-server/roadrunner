package memory

import (
	"sync"

	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	PluginName string = "memory"
)

type Plugin struct {
	log logger.Logger

	// channel with the messages from the RPC
	pushCh chan []byte
	// user-subscribed topics
	topics sync.Map
}

func (p *Plugin) Init(log logger.Logger) error {
	p.log = log
	p.pushCh = make(chan []byte, 100)
	return nil
}

// Available interface implementation for the plugin
func (p *Plugin) Available() {}

// Name is endure.Named interface implementation
func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Publish(messages []byte) error {
	p.pushCh <- messages
	return nil
}

func (p *Plugin) PublishAsync(messages []byte) {
	go func() {
		p.pushCh <- messages
	}()
}

func (p *Plugin) Subscribe(topics ...string) error {
	for i := 0; i < len(topics); i++ {
		p.topics.Store(topics[i], struct{}{})
	}
	return nil
}

func (p *Plugin) Unsubscribe(topics ...string) error {
	for i := 0; i < len(topics); i++ {
		p.topics.Delete(topics[i])
	}
	return nil
}

func (p *Plugin) Next() (*pubsub.Message, error) {
	msg := <-p.pushCh

	if msg == nil {
		return nil, nil
	}

	// push only messages, which are subscribed
	// TODO better???
	for i := 0; i < len(msg.Topics); i++ {
		if _, ok := p.topics.Load(msg.Topics[i]); ok {
			return msg, nil
		}
	}
	return nil, nil
}
