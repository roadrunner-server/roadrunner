package memory

import (
	"sync"

	"github.com/spiral/roadrunner/v2/pkg/bst"
	"github.com/spiral/roadrunner/v2/pkg/pubsub/message"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"google.golang.org/protobuf/proto"
)

const (
	PluginName string = "memory"
)

type Plugin struct {
	sync.RWMutex
	log logger.Logger

	// channel with the messages from the RPC
	pushCh chan []byte
	// user-subscribed topics
	storage bst.Storage
}

func (p *Plugin) Init(log logger.Logger) error {
	p.log = log
	p.pushCh = make(chan []byte, 10)
	p.storage = bst.NewBST()
	return nil
}

// Available interface implementation for the plugin
func (p *Plugin) Available() {}

// Name is endure.Named interface implementation
func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Publish(message []byte) error {
	p.pushCh <- message
	return nil
}

func (p *Plugin) PublishAsync(message []byte) {
	go func() {
		p.pushCh <- message
	}()
}

func (p *Plugin) Subscribe(connectionID string, topics ...string) error {
	p.Lock()
	defer p.Unlock()
	for i := 0; i < len(topics); i++ {
		p.storage.Insert(connectionID, topics[i])
	}
	return nil
}

func (p *Plugin) Unsubscribe(connectionID string, topics ...string) error {
	p.Lock()
	defer p.Unlock()
	for i := 0; i < len(topics); i++ {
		p.storage.Remove(connectionID, topics[i])
	}
	return nil
}

func (p *Plugin) Connections(topic string, res map[string]struct{}) {
	p.RLock()
	defer p.RUnlock()

	ret := p.storage.Get(topic)
	for rr := range ret {
		res[rr] = struct{}{}
	}
}

func (p *Plugin) Next() (*message.Message, error) {
	msg := <-p.pushCh
	if msg == nil {
		return nil, nil
	}

	p.RLock()
	defer p.RUnlock()

	m := &message.Message{}
	err := proto.Unmarshal(msg, m)
	if err != nil {
		return nil, err
	}

	// push only messages, which are subscribed
	// TODO better???
	for i := 0; i < len(m.GetTopics()); i++ {
		// if we have active subscribers - send a message to a topic
		// or send nil instead
		if ok := p.storage.Contains(m.GetTopics()[i]); ok {
			return m, nil
		}
	}
	return nil, nil
}
