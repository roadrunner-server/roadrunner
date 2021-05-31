package executor

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/fasthttp/websocket"
	json "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/websockets/commands"
	"github.com/spiral/roadrunner/v2/plugins/websockets/connection"
	"github.com/spiral/roadrunner/v2/plugins/websockets/storage"
	"github.com/spiral/roadrunner/v2/plugins/websockets/validator"
)

type Response struct {
	Topic   string   `json:"topic"`
	Payload []string `json:"payload"`
}

type Executor struct {
	sync.Mutex
	conn    *connection.Connection
	storage *storage.Storage
	log     logger.Logger

	// associated connection ID
	connID string

	// map with the pubsub drivers
	pubsub       map[string]pubsub.PubSub
	actualTopics map[string]struct{}

	req             *http.Request
	accessValidator validator.AccessValidatorFn
}

// NewExecutor creates protected connection and starts command loop
func NewExecutor(conn *connection.Connection, log logger.Logger, bst *storage.Storage,
	connID string, pubsubs map[string]pubsub.PubSub, av validator.AccessValidatorFn, r *http.Request) *Executor {
	return &Executor{
		conn:            conn,
		connID:          connID,
		storage:         bst,
		log:             log,
		pubsub:          pubsubs,
		accessValidator: av,
		actualTopics:    make(map[string]struct{}, 10),
		req:             r,
	}
}

func (e *Executor) StartCommandLoop() error { //nolint:gocognit
	const op = errors.Op("executor_command_loop")
	for {
		mt, data, err := e.conn.Read()
		if err != nil {
			if mt == -1 {
				e.log.Info("socket was closed", "reason", err, "message type", mt)
				return nil
			}

			return errors.E(op, err)
		}

		msg := &pubsub.Message{}

		err = json.Unmarshal(data, msg)
		if err != nil {
			e.log.Error("error unmarshal message", "error", err)
			continue
		}

		// nil message, continue
		if msg == nil {
			e.log.Warn("get nil message, skipping")
			continue
		}

		switch msg.Command {
		// handle leave
		case commands.Join:
			e.log.Debug("get join command", "msg", msg)

			val, err := e.accessValidator(e.req, msg.Topics...)
			if err != nil {
				if val != nil {
					e.log.Debug("validation error", "status", val.Status, "headers", val.Header, "body", val.Body)
				}

				resp := &Response{
					Topic:   "#join",
					Payload: msg.Topics,
				}

				packet, errJ := json.Marshal(resp)
				if errJ != nil {
					e.log.Error("error marshal the body", "error", errJ)
					return errors.E(op, fmt.Errorf("%v,%v", err, errJ))
				}

				errW := e.conn.Write(websocket.BinaryMessage, packet)
				if errW != nil {
					e.log.Error("error writing payload to the connection", "payload", packet, "error", errW)
					return errors.E(op, fmt.Errorf("%v,%v", err, errW))
				}

				continue
			}

			resp := &Response{
				Topic:   "@join",
				Payload: msg.Topics,
			}

			packet, err := json.Marshal(resp)
			if err != nil {
				e.log.Error("error marshal the body", "error", err)
				return errors.E(op, err)
			}

			err = e.conn.Write(websocket.BinaryMessage, packet)
			if err != nil {
				e.log.Error("error writing payload to the connection", "payload", packet, "error", err)
				return errors.E(op, err)
			}

			// subscribe to the topic
			if br, ok := e.pubsub[msg.Broker]; ok {
				err = e.Set(br, msg.Topics)
				if err != nil {
					return errors.E(op, err)
				}
			}

		// handle leave
		case commands.Leave:
			e.log.Debug("get leave command", "msg", msg)

			// prepare response
			resp := &Response{
				Topic:   "@leave",
				Payload: msg.Topics,
			}

			packet, err := json.Marshal(resp)
			if err != nil {
				e.log.Error("error marshal the body", "error", err)
				return errors.E(op, err)
			}

			err = e.conn.Write(websocket.BinaryMessage, packet)
			if err != nil {
				e.log.Error("error writing payload to the connection", "payload", packet, "error", err)
				return errors.E(op, err)
			}

			if br, ok := e.pubsub[msg.Broker]; ok {
				err = e.Leave(br, msg.Topics)
				if err != nil {
					return errors.E(op, err)
				}
			}

		case commands.Headers:

		default:
			e.log.Warn("unknown command", "command", msg.Command)
		}
	}
}

func (e *Executor) Set(br pubsub.PubSub, topics []string) error {
	// associate connection with topics
	err := br.Subscribe(topics...)
	if err != nil {
		e.log.Error("error subscribing to the provided topics", "topics", topics, "error", err.Error())
		// in case of error, unsubscribe connection from the dead topics
		_ = br.Unsubscribe(topics...)
		return err
	}

	e.storage.InsertMany(e.connID, topics)

	// save topics for the connection
	for i := 0; i < len(topics); i++ {
		e.actualTopics[topics[i]] = struct{}{}
	}

	return nil
}

func (e *Executor) Leave(br pubsub.PubSub, topics []string) error {
	// remove associated connections from the storage
	e.storage.RemoveMany(e.connID, topics)
	err := br.Unsubscribe(topics...)
	if err != nil {
		e.log.Error("error subscribing to the provided topics", "topics", topics, "error", err.Error())
		return err
	}

	// remove topics for the connection
	for i := 0; i < len(topics); i++ {
		delete(e.actualTopics, topics[i])
	}

	return nil
}

func (e *Executor) CleanUp() {
	for topic := range e.actualTopics {
		// remove from the bst
		e.storage.Remove(e.connID, topic)

		for _, ps := range e.pubsub {
			_ = ps.Unsubscribe(topic)
		}
	}

	for k := range e.actualTopics {
		delete(e.actualTopics, k)
	}
}
