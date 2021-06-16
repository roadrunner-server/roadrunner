package pool

import (
	"sync"

	json "github.com/json-iterator/go"
	websocketsv1 "github.com/spiral/roadrunner/v2/pkg/proto/websockets/v1beta"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/websockets/connection"
	"github.com/spiral/roadrunner/v2/utils"
)

type WorkersPool struct {
	storage     map[string]pubsub.PubSub
	connections *sync.Map
	resPool     sync.Pool
	log         logger.Logger

	queue chan *websocketsv1.Message
	exit  chan struct{}
}

// NewWorkersPool constructs worker pool for the websocket connections
func NewWorkersPool(pubsubs map[string]pubsub.PubSub, connections *sync.Map, log logger.Logger) *WorkersPool {
	wp := &WorkersPool{
		connections: connections,
		queue:       make(chan *websocketsv1.Message, 100),
		storage:     pubsubs,
		log:         log,
		exit:        make(chan struct{}),
	}

	wp.resPool.New = func() interface{} {
		return make(map[string]struct{}, 10)
	}

	// start 10 workers
	for i := 0; i < 50; i++ {
		wp.do()
	}

	return wp
}

func (wp *WorkersPool) Queue(msg *websocketsv1.Message) {
	wp.queue <- msg
}

func (wp *WorkersPool) Stop() {
	for i := 0; i < 50; i++ {
		wp.exit <- struct{}{}
	}

	close(wp.exit)
}

func (wp *WorkersPool) put(res map[string]struct{}) {
	// optimized
	// https://go-review.googlesource.com/c/go/+/110055/
	// not O(n), but O(1)
	for k := range res {
		delete(res, k)
	}
}

func (wp *WorkersPool) get() map[string]struct{} {
	return wp.resPool.Get().(map[string]struct{})
}

// Response from the server
type Response struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

func (wp *WorkersPool) do() { //nolint:gocognit
	go func() {
		for {
			select {
			case msg, ok := <-wp.queue:
				if !ok {
					return
				}
				_ = msg
				if msg == nil {
					continue
				}
				if len(msg.GetTopics()) == 0 {
					continue
				}

				br, ok := wp.storage[msg.Broker]
				if !ok {
					wp.log.Warn("no such broker", "requested", msg.GetBroker(), "available", wp.storage)
					continue
				}

				// send a message to every topic
				for i := 0; i < len(msg.GetTopics()); i++ {
					// get free map
					res := wp.get()

					// get connections for the particular topic
					br.Connections(msg.GetTopics()[i], res)

					if len(res) == 0 {
						wp.log.Info("no such topic", "topic", msg.GetTopics()[i])
						wp.put(res)
						continue
					}

					// res is a map with a connectionsID
					for topic := range res {
						c, ok := wp.connections.Load(topic)
						if !ok {
							wp.log.Warn("the user disconnected connection before the message being written to it", "broker", msg.GetBroker(), "topics", msg.GetTopics()[i])
							wp.put(res)
							continue
						}

						response := &Response{
							Topic:   msg.GetTopics()[i],
							Payload: utils.AsString(msg.GetPayload()),
						}

						d, err := json.Marshal(response)
						if err != nil {
							wp.log.Error("error marshaling response", "error", err)
							wp.put(res)
							break
						}

						// put data into the bytes buffer
						err = c.(*connection.Connection).Write(d)
						if err != nil {
							for i := 0; i < len(msg.GetTopics()); i++ {
								wp.log.Error("error sending payload over the connection", "error", err, "broker", msg.GetBroker(), "topics", msg.GetTopics()[i])
							}
							wp.put(res)
							continue
						}
					}
				}
			case <-wp.exit:
				wp.log.Info("get exit signal, exiting from the workers pool")
				return
			}
		}
	}()
}
