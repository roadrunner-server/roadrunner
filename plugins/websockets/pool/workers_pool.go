package pool

import (
	"sync"

	"github.com/fasthttp/websocket"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/pkg/pubsub/message"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/websockets/connection"
)

type WorkersPool struct {
	storage     map[string]pubsub.PubSub
	connections *sync.Map
	resPool     sync.Pool
	log         logger.Logger

	queue chan *message.Message
	exit  chan struct{}
}

// NewWorkersPool constructs worker pool for the websocket connections
func NewWorkersPool(pubsubs map[string]pubsub.PubSub, connections *sync.Map, log logger.Logger) *WorkersPool {
	wp := &WorkersPool{
		connections: connections,
		queue:       make(chan *message.Message, 100),
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

func (wp *WorkersPool) Queue(msg *message.Message) {
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

				res := wp.get()

				for i := 0; i < len(msg.GetTopics()); i++ {
					// get connections for the particular topic
					br.Connections(msg.GetTopics()[i], res)
				}

				if len(res) == 0 {
					for i := 0; i < len(msg.GetTopics()); i++ {
						wp.log.Info("no such topic", "topic", msg.GetTopics()[i])
					}
					wp.put(res)
					continue
				}

				for i := range res {
					c, ok := wp.connections.Load(i)
					if !ok {
						for i := 0; i < len(msg.GetTopics()); i++ {
							wp.log.Warn("the user disconnected connection before the message being written to it", "broker", msg.GetBroker(), "topics", msg.GetTopics()[i])
						}
						continue
					}

					conn := c.(*connection.Connection)

					// put data into the bytes buffer
					err := conn.Write(websocket.BinaryMessage, msg.GetPayload())
					if err != nil {
						for i := 0; i < len(msg.GetTopics()); i++ {
							wp.log.Error("error sending payload over the connection", "error", err, "broker", msg.GetBroker(), "topics", msg.GetTopics()[i])
						}
						continue
					}
				}

				// put map with results back
				wp.put(res)
			case <-wp.exit:
				wp.log.Info("get exit signal, exiting from the workers pool")
				return
			}
		}
	}()
}
