package pubsub

import "github.com/spiral/roadrunner/v2/pkg/pubsub/message"

// PubSub ...
type PubSub interface {
	Publisher
	Subscriber
	Reader
}

// Subscriber defines the ability to operate as message passing broker.
type Subscriber interface {
	// Subscribe broker to one or multiple topics.
	Subscribe(topics ...string) error

	// Unsubscribe from one or multiply topics
	Unsubscribe(topics ...string) error
}

// Publisher publish one or more messages
type Publisher interface {
	// Publish one or multiple Channel.
	Publish(messages []byte) error

	// PublishAsync publish message and return immediately
	// If error occurred it will be printed into the logger
	PublishAsync(messages []byte)
}

// Reader interface should return next message
type Reader interface {
	Next() (*message.Message, error)
}
