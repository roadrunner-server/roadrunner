package pubsub

type Message struct {
	// Command (join, leave, headers)
	Command string `json:"command"`

	// Broker (redis, memory)
	Broker string `json:"broker"`

	// Topic message been pushed into.
	Topics []string `json:"topic"`

	// Payload to be broadcasted
	Payload []byte `json:"payload"`
}
