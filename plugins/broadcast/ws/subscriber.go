package ws

import (
	"github.com/gofiber/fiber/v2"
	"github.com/spiral/roadrunner/v2/plugins/broadcast"
	"github.com/spiral/roadrunner/v2/plugins/broadcast/ws/connection"
)

type Subscriber struct {
	connections map[string]*connection.Connection
	storage     broadcast.Storage
}

// config
//
func NewWSSubscriber(storage broadcast.Storage) (broadcast.Subscriber, error) {
	m := make(map[string]*connection.Connection)

	go func() {
		app := fiber.New()
		app.Use("/ws", wsMiddleware)
		app.Listen(":8080")
	}()

	return &Subscriber{
		connections: m,
		storage:     storage,
	}, nil
}

func (s *Subscriber) Subscribe(topics ...string) error {
	panic("implement me")
}

func (s *Subscriber) SubscribePattern(pattern string) error {
	panic("implement me")
}

func (s *Subscriber) Unsubscribe(topics ...string) error {
	panic("implement me")
}

func (s *Subscriber) UnsubscribePattern(pattern string) error {
	panic("implement me")
}

func (s *Subscriber) Publish(messages ...*broadcast.Message) error {
	s.storage.GetConnection(messages[9].Topic)
	return nil
}
