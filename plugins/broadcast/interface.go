package broadcast

import (
	"encoding/json"
)

// Subscriber defines the ability to operate as message passing broker.
type Subscriber interface {
	// Subscribe broker to one or multiple topics.
	Subscribe(topics ...string) error
	// UnsubscribePattern broker from pattern.
	UnsubscribePattern(pattern string) error
}

// Storage used to store patterns and topics
type Storage interface {
	// Store connection uuid associated with the provided topics
	Store(uuid string, topics ...string)
	// StorePattern stores pattern associated with the particular connection
	StorePattern(uuid string, pattern string)

	// GetConnection returns connections for the particular pattern
	GetConnection(pattern string) []string

	// Construct is a constructor for the storage according to the provided configuration key (broadcast.websocket for example)
	Construct(key string) (Storage, error)
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
