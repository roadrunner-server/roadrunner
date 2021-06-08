package memory

import (
	"sync"

	"github.com/spiral/roadrunner/v2/pkg/bst"
	websocketsv1 "github.com/spiral/roadrunner/v2/pkg/proto/websockets/v1beta"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"google.golang.org/protobuf/proto"
)

type Plugin struct {
	sync.RWMutex
	log logger.Logger

	// channel with the messages from the RPC
	pushCh chan []byte
	// user-subscribed topics
	storage bst.Storage
}

func NewInMemory(log logger.Logger) pubsub.PubSub {
	return &Plugin{
		log:     log,
		pushCh:  make(chan []byte, 10),
		storage: bst.NewBST(),
	}
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

func (p *Plugin) Next() (*websocketsv1.Message, error) {
	msg := <-p.pushCh
	if msg == nil {
		return nil, nil
	}

	p.RLock()
	defer p.RUnlock()

	m := &websocketsv1.Message{}
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
