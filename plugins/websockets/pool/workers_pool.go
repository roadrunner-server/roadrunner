package pool

import (
	"sync"

	json "github.com/json-iterator/go"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/websockets/connection"
	"github.com/spiral/roadrunner/v2/utils"
)

type WorkersPool struct {
	subscriber  pubsub.Subscriber
	connections *sync.Map
	resPool     sync.Pool
	log         logger.Logger

	queue chan *pubsub.Message
	exit  chan struct{}
}

// NewWorkersPool constructs worker pool for the websocket connections
func NewWorkersPool(subscriber pubsub.Subscriber, connections *sync.Map, log logger.Logger) *WorkersPool {
	wp := &WorkersPool{
		connections: connections,
		queue:       make(chan *pubsub.Message, 100),
		subscriber:  subscriber,
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

func (wp *WorkersPool) Queue(msg *pubsub.Message) {
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
				if msg == nil || msg.Topic == "" {
					continue
				}

				// get free map
				res := wp.get()

				// get connections for the particular topic
				wp.subscriber.Connections(msg.Topic, res)

				if len(res) == 0 {
					wp.log.Info("no connections associated with provided topic", "topic", msg.Topic)
					wp.put(res)
					continue
				}

				// res is a map with a connectionsID
				for connID := range res {
					c, ok := wp.connections.Load(connID)
					if !ok {
						wp.log.Warn("the websocket disconnected before the message being written to it", "topics", msg.Topic)
						wp.put(res)
						continue
					}

					d, err := json.Marshal(&Response{
						Topic:   msg.Topic,
						Payload: utils.AsString(msg.Payload),
					})

					if err != nil {
						wp.log.Error("error marshaling response", "error", err)
						wp.put(res)
						break
					}

					// put data into the bytes buffer
					err = c.(*connection.Connection).Write(d)
					if err != nil {
						wp.log.Error("error sending payload over the connection", "error", err, "topic", msg.Topic)
						wp.put(res)
						continue
					}
				}
			case <-wp.exit:
				wp.log.Info("get exit signal, exiting from the workers pool")
				return
			}
		}
	}()
}
