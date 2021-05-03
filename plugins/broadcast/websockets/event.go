package websockets

import (
	"github.com/gorilla/websocket"
)

const (
	// EventConnect fired when new client is connected, the context is *websocket.Conn.
	EventConnect = iota + 2500

	// EventDisconnect fired when websocket is disconnected, context is empty.
	EventDisconnect

	// EventJoin caused when topics are being consumed, context if *TopicEvent.
	EventJoin

	// EventLeave caused when topic consumption are stopped, context if *TopicEvent.
	EventLeave

	// EventError when any broadcast error occurred, the context is *ErrorEvent.
	EventError
)

// ErrorEvent represents singular broadcast error event.
type ErrorEvent struct {
	// Conn specific to the error.
	Conn *websocket.Conn

	// Error contains job specific error.
	Error error
}

// TopicEvent caused when topic is joined or left.
type TopicEvent struct {
	// Conn associated with topics.
	Conn *websocket.Conn

	// Topics specific to event.
	Topics []string
}
