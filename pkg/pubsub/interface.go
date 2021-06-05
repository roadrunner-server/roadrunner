package pubsub

import "github.com/spiral/roadrunner/v2/pkg/pubsub/message"

/*
This interface is in BETA. It might be changed.
*/

// PubSub ...
type PubSub interface {
	Publisher
	Subscriber
	Reader
}

// Subscriber defines the ability to operate as message passing broker.
// BETA interface
type Subscriber interface {
	// Subscribe broker to one or multiple topics.
	Subscribe(connectionID string, topics ...string) error

	// Unsubscribe from one or multiply topics
	Unsubscribe(connectionID string, topics ...string) error

	// Connections returns all connections associated with the particular topic
	Connections(topic string, ret map[string]struct{})
}

// Publisher publish one or more messages
// BETA interface
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
