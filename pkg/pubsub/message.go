package pubsub

import (
	json "github.com/json-iterator/go"
)

type Message struct {
	// Topic message been pushed into.
	Topics []string `json:"topic"`

	// Command (join, leave, headers)
	Command string `json:"command"`

	// Broker (redis, memory)
	Broker string `json:"broker"`

	// Payload to be broadcasted
	Payload []byte `json:"payload"`
}

// MarshalBinary needed to marshal message for the redis
func (m *Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}
