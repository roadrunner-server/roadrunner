package pubsub

import json "github.com/json-iterator/go"

// Message represents a single message with payload bound to a particular topic
type Message struct {
	// Topic (channel in terms of redis)
	Topic string `json:"topic"`
	// Payload (on some decode stages might be represented as base64 string)
	Payload []byte `json:"payload"`
}

func (m *Message) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}
