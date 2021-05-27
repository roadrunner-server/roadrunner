package executor

import (
	"github.com/fasthttp/websocket"
	json "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/websockets/commands"
	"github.com/spiral/roadrunner/v2/plugins/websockets/connection"
	"github.com/spiral/roadrunner/v2/plugins/websockets/storage"
)

type Response struct {
	Topic   string   `json:"topic"`
	Payload []string `json:"payload"`
}

type Executor struct {
	conn    *connection.Connection
	storage *storage.Storage
	log     logger.Logger

	// associated connection ID
	connID string

	// map with the pubsub drivers
	pubsub map[string]pubsub.PubSub
}

// NewExecutor creates protected connection and starts command loop
func NewExecutor(conn *connection.Connection, log logger.Logger, bst *storage.Storage, connID string, pubsubs map[string]pubsub.PubSub) *Executor {
	return &Executor{
		conn:    conn,
		connID:  connID,
		storage: bst,
		log:     log,
		pubsub:  pubsubs,
	}
}

func (e *Executor) StartCommandLoop() error {
	const op = errors.Op("executor_command_loop")
	for {
		mt, data, err := e.conn.Read()
		if err != nil {
			if mt == -1 {
				e.log.Error("socket was closed", "error", err, "message type", mt)
				return nil
			}

			return errors.E(op, err)
		}

		msg := &pubsub.Msg{}

		err = json.Unmarshal(data, msg)
		if err != nil {
			e.log.Error("error unmarshal message", "error", err)
			continue
		}

		switch msg.Command() {
		// handle leave
		case commands.Join:
			// associate connection with topics
			e.storage.Store(e.connID, msg.Topics())

			resp := &Response{
				Topic:   "@join",
				Payload: msg.Topics(),
			}

			packet, err := json.Marshal(resp)
			if err != nil {
				e.log.Error("error marshal the body", "error", err)
				continue
			}

			err = e.conn.Write(websocket.BinaryMessage, packet)
			if err != nil {
				e.log.Error("error writing payload to the connection", "payload", packet, "error", err)
				continue
			}

			err = e.pubsub[msg.Broker()].Subscribe(msg.Topics()...)
			if err != nil {
				e.log.Error("error subscribing to the provided topics", "topics", msg.Topics(), "error", err.Error())
				continue
			}

		// handle leave
		case commands.Leave:
			// remove associated connections from the storage
			e.storage.Remove(e.connID, msg.Topics())

			resp := &Response{
				Topic:   "@leave",
				Payload: msg.Topics(),
			}

			packet, err := json.Marshal(resp)
			if err != nil {
				e.log.Error("error marshal the body", "error", err)
				continue
			}

			err = e.pubsub[msg.Broker()].Unsubscribe(msg.Topics()...)
			if err != nil {
				e.log.Error("error subscribing to the provided topics", "topics", msg.Topics(), "error", err.Error())
				continue
			}

			err = e.conn.Write(websocket.BinaryMessage, packet)
			if err != nil {
				e.log.Error("error writing payload to the connection", "payload", packet, "error", err)
				continue
			}

		case commands.Headers:

		default:
			e.log.Warn("unknown command", "command", msg.Command())
		}
	}
}
