package ws

import "github.com/spiral/roadrunner/v2/plugins/broadcast"

type Subscriber struct {
	connections map[string]*Connection
	storage broadcast.Storage
}

func NewWSSubscriber() (broadcast.Subscriber, error) {
	m := make(map[string]*Connection)
	return &Subscriber{
		connections: m,
	}, nil
}

func (s *Subscriber) Subscribe(upstream chan *broadcast.Message, topics ...string) error {
	panic("implement me")




}

func (s *Subscriber) SubscribePattern(upstream chan *broadcast.Message, pattern string) error {
	panic("implement me")
}

func (s *Subscriber) Unsubscribe(upstream chan *broadcast.Message, topics ...string) error {
	panic("implement me")
}

func (s *Subscriber) UnsubscribePattern(upstream chan *broadcast.Message, pattern string) error {
	panic("implement me")
}
