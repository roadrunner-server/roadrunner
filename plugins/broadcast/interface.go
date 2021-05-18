package broadcast

import "encoding/json"

// Subscriber defines the ability to operate as message passing broker.
type Subscriber interface {
	// Subscribe broker to one or multiple topics.
	Subscribe(upstream chan *Message, topics ...string) error

	// SubscribePattern broker to pattern.
	SubscribePattern(upstream chan *Message, pattern string) error

	// Unsubscribe broker from one or multiple topics.
	Unsubscribe(upstream chan *Message, topics ...string) error

	// UnsubscribePattern broker from pattern.
	UnsubscribePattern(upstream chan *Message, pattern string) error
}

type Storage interface {
	Store(topics ...string)
	StorePattern(pattern string)
}

type Publisher interface {
	// Publish one or multiple Channel.
	Publish(messages ...*Message) error
}

// Message represent single message.
type Message struct {
	// Topic message been pushed into.
	Topic string `json:"topic"`

	// Payload to be broadcasted. Must be valid json when transferred over RPC.
	Payload json.RawMessage `json:"payload"`
}
