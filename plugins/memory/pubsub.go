package memory

import (
	"sync"

	"github.com/spiral/roadrunner/v2/pkg/bst"
	websocketsv1 "github.com/spiral/roadrunner/v2/pkg/proto/websockets/v1beta"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"google.golang.org/protobuf/proto"
)

type PubSubDriver struct {
	sync.RWMutex
	// channel with the messages from the RPC
	pushCh chan []byte
	// user-subscribed topics
	storage bst.Storage
	log     logger.Logger
}

func NewPubSubDriver(log logger.Logger, _ string) (pubsub.PubSub, error) {
	ps := &PubSubDriver{
		pushCh:  make(chan []byte, 10),
		storage: bst.NewBST(),
		log:     log,
	}
	return ps, nil
}

func (p *PubSubDriver) Publish(message []byte) error {
	p.pushCh <- message
	return nil
}

func (p *PubSubDriver) PublishAsync(message []byte) {
	go func() {
		p.pushCh <- message
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

func (p *PubSubDriver) Next() (*websocketsv1.Message, error) {
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
