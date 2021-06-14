package pubsub

import websocketsv1 "github.com/spiral/roadrunner/v2/pkg/proto/websockets/v1beta"

/*
This interface is in BETA. It might be changed.
*/

// PubSub interface designed to implement on any storage type to provide pub-sub abilities
// Publisher used to receive messages from the PHP app via RPC
// Subscriber should be implemented to subscribe to a topics and provide a connections list per topic
// Reader return next message from the channel
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
	Next() (*websocketsv1.Message, error)
}

type PSProvider interface {
	PSProvide(key string) (PubSub, error)
}
