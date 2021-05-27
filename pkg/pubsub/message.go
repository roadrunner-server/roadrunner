package pubsub

import (
	json "github.com/json-iterator/go"
)

type Msg struct {
	// Topic message been pushed into.
	Topics_ []string `json:"topic"`

	// Command (join, leave, headers)
	Command_ string `json:"command"`

	// Broker (redis, memory)
	Broker_ string `json:"broker"`

	// Payload to be broadcasted
	Payload_ []byte `json:"payload"`
}

func (m *Msg) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

// Payload in raw bytes
func (m *Msg) Payload() []byte {
	return m.Payload_
}

// Command for the connection
func (m *Msg) Command() string {
	return m.Command_
}

// Topics to subscribe
func (m *Msg) Topics() []string {
	return m.Topics_
}

func (m *Msg) Broker() string {
	return m.Broker_
}
