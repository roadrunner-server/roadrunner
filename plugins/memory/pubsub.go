package memory

import (
	"sync"

	"github.com/spiral/roadrunner/v2/pkg/bst"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type PubSubDriver struct {
	sync.RWMutex
	// channel with the messages from the RPC
	pushCh chan *pubsub.Message
	// user-subscribed topics
	storage bst.Storage
	log     logger.Logger
}

func NewPubSubDriver(log logger.Logger, _ string) (pubsub.PubSub, error) {
	ps := &PubSubDriver{
		pushCh:  make(chan *pubsub.Message, 10),
		storage: bst.NewBST(),
		log:     log,
	}
	return ps, nil
}

func (p *PubSubDriver) Publish(msg *pubsub.Message) error {
	p.pushCh <- msg
	return nil
}

func (p *PubSubDriver) PublishAsync(msg *pubsub.Message) {
	go func() {
		p.pushCh <- msg
	}()
}

func (p *PubSubDriver) Subscribe(connectionID string, topics ...string) error {
	p.Lock()
	defer p.Unlock()
	for i := 0; i < len(topics); i++ {
		p.storage.Insert(connectionID, topics[i])
	}
	return nil
}

func (p *PubSubDriver) Unsubscribe(connectionID string, topics ...string) error {
	p.Lock()
	defer p.Unlock()
	for i := 0; i < len(topics); i++ {
		p.storage.Remove(connectionID, topics[i])
	}
	return nil
}

func (p *PubSubDriver) Connections(topic string, res map[string]struct{}) {
	p.RLock()
	defer p.RUnlock()

	ret := p.storage.Get(topic)
	for rr := range ret {
		res[rr] = struct{}{}
	}
}

func (p *PubSubDriver) Next() (*pubsub.Message, error) {
	msg := <-p.pushCh
	if msg == nil {
		return nil, nil
	}

	p.RLock()
	defer p.RUnlock()

	// push only messages, which topics are subscibed
	// TODO better???
	// if we have active subscribers - send a message to a topic
	// or send nil instead
	if ok := p.storage.Contains(msg.Topic); ok {
		return msg, nil
	}

	return nil, nil
}
