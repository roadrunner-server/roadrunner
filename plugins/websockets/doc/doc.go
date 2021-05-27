package doc

/*
RPC message structure:

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

1. Topics - string array (slice) with topics to join or leave
2. Command - string, command to apply on the provided topics
3. Broker - string, pub-sub broker to use, for the one-node systems might be used `memory` broker or `redis`. For the multi-node -
`redis` broker should be used.
4. Payload - raw byte array to send to the subscribers (binary messages).


*/
