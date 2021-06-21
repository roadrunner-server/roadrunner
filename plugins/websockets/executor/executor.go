package executor

import (
	"fmt"
	"net/http"
	"sync"

	json "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/websockets/commands"
	"github.com/spiral/roadrunner/v2/plugins/websockets/connection"
	"github.com/spiral/roadrunner/v2/plugins/websockets/validator"
	websocketsv1 "github.com/spiral/roadrunner/v2/proto/websockets/v1beta"
)

type Response struct {
	Topic   string   `json:"topic"`
	Payload []string `json:"payload"`
}

type Executor struct {
	sync.Mutex
	// raw ws connection
	conn *connection.Connection
	log  logger.Logger

	// associated connection ID
	connID string

	// subscriber drivers
	sub          pubsub.Subscriber
	actualTopics map[string]struct{}

	req             *http.Request
	accessValidator validator.AccessValidatorFn
}

// NewExecutor creates protected connection and starts command loop
func NewExecutor(conn *connection.Connection, log logger.Logger,
	connID string, sub pubsub.Subscriber, av validator.AccessValidatorFn, r *http.Request) *Executor {
	return &Executor{
		conn:            conn,
		connID:          connID,
		log:             log,
		sub:             sub,
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

		msg := &websocketsv1.Message{}

		err = json.Unmarshal(data, msg)
		if err != nil {
			e.log.Error("unmarshal message", "error", err)
			continue
		}

		// nil message, continue
		if msg == nil {
			e.log.Warn("nil message, skipping")
			continue
		}

		switch msg.Command {
		// handle leave
		case commands.Join:
			e.log.Debug("received join command", "msg", msg)

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
					e.log.Error("marshal the body", "error", errJ)
					return errors.E(op, fmt.Errorf("%v,%v", err, errJ))
				}

				errW := e.conn.Write(packet)
				if errW != nil {
					e.log.Error("write payload to the connection", "payload", packet, "error", errW)
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
				e.log.Error("marshal the body", "error", err)
				return errors.E(op, err)
			}

			err = e.conn.Write(packet)
			if err != nil {
				e.log.Error("write payload to the connection", "payload", packet, "error", err)
				return errors.E(op, err)
			}

			// subscribe to the topic
			err = e.Set(msg.Topics)
			if err != nil {
				return errors.E(op, err)
			}

		// handle leave
		case commands.Leave:
			e.log.Debug("received leave command", "msg", msg)

			// prepare response
			resp := &Response{
				Topic:   "@leave",
				Payload: msg.Topics,
			}

			packet, err := json.Marshal(resp)
			if err != nil {
				e.log.Error("marshal the body", "error", err)
				return errors.E(op, err)
			}

			err = e.conn.Write(packet)
			if err != nil {
				e.log.Error("write payload to the connection", "payload", packet, "error", err)
				return errors.E(op, err)
			}

			err = e.Leave(msg.Topics)
			if err != nil {
				return errors.E(op, err)
			}

		case commands.Headers:

		default:
			e.log.Warn("unknown command", "command", msg.Command)
		}
	}
}

func (e *Executor) Set(topics []string) error {
	// associate connection with topics
	err := e.sub.Subscribe(e.connID, topics...)
	if err != nil {
		e.log.Error("subscribe to the provided topics", "topics", topics, "error", err.Error())
		// in case of error, unsubscribe connection from the dead topics
		_ = e.sub.Unsubscribe(e.connID, topics...)
		return err
	}

	// save topics for the connection
	for i := 0; i < len(topics); i++ {
		e.actualTopics[topics[i]] = struct{}{}
	}

	return nil
}

func (e *Executor) Leave(topics []string) error {
	// remove associated connections from the storage
	err := e.sub.Unsubscribe(e.connID, topics...)
	if err != nil {
		e.log.Error("subscribe to the provided topics", "topics", topics, "error", err.Error())
		return err
	}

	// remove topics for the connection
	for i := 0; i < len(topics); i++ {
		delete(e.actualTopics, topics[i])
	}

	return nil
}

func (e *Executor) CleanUp() {
	// unsubscribe particular connection from the topics
	for topic := range e.actualTopics {
		_ = e.sub.Unsubscribe(e.connID, topic)
	}

	// clean up the actualTopics data
	for k := range e.actualTopics {
		delete(e.actualTopics, k)
	}
}
