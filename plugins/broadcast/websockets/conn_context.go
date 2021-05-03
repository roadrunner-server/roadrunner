package websockets

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

// ConnContext carries information about websocket connection and it's topics.
type ConnContext struct {
	// Conn to the client.
	Conn *websocket.Conn

	// Topics contain list of currently subscribed topics.
	Topics []string

	// upstream to push messages into.
	upstream chan *broadcast.Message
}

// SendMessage message directly to the client.
func (ctx *ConnContext) SendMessage(topic string, payload interface{}) (err error) {
	msg := &broadcast.Message{Topic: topic}
	msg.Payload, err = json.Marshal(payload)

	if err == nil {
		ctx.upstream <- msg
	}

	return err
}

func (ctx *ConnContext) serve(errHandler func(err error, conn *websocket.Conn)) {
	for msg := range ctx.upstream {
		if err := ctx.Conn.WriteJSON(msg); err != nil {
			errHandler(err, ctx.Conn)
		}
	}
}

func (ctx *ConnContext) addTopics(topics ...string) {
	for _, topic := range topics {
		found := false
		for _, e := range ctx.Topics {
			if e == topic {
				found = true
				break
			}
		}

		if !found {
			ctx.Topics = append(ctx.Topics, topic)
		}
	}
}

func (ctx *ConnContext) dropTopic(topics ...string) {
	for _, topic := range topics {
		for i, e := range ctx.Topics {
			if e == topic {
				ctx.Topics[i] = ctx.Topics[len(ctx.Topics)-1]
				ctx.Topics = ctx.Topics[:len(ctx.Topics)-1]
			}
		}
	}
}
