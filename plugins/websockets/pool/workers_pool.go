package pool

import (
	"sync"

	"github.com/fasthttp/websocket"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/websockets/connection"
	"github.com/spiral/roadrunner/v2/plugins/websockets/storage"
)

type WorkersPool struct {
	storage     *storage.Storage
	connections *sync.Map
	resPool     sync.Pool
	log         logger.Logger

	queue chan *pubsub.Message
	exit  chan struct{}
}

// NewWorkersPool constructs worker pool for the websocket connections
func NewWorkersPool(storage *storage.Storage, connections *sync.Map, log logger.Logger) *WorkersPool {
	wp := &WorkersPool{
		connections: connections,
		queue:       make(chan *pubsub.Message, 100),
		storage:     storage,
		log:         log,
		exit:        make(chan struct{}),
	}

	wp.resPool.New = func() interface{} {
		return make(map[string]struct{}, 10)
	}

	// start 10 workers
	for i := 0; i < 10; i++ {
		wp.do()
	}

	return wp
}

func (wp *WorkersPool) Queue(msg *pubsub.Message) {
	wp.queue <- msg
}

func (wp *WorkersPool) Stop() {
	for i := 0; i < 10; i++ {
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
				// do not handle nil's
				if msg == nil {
					continue
				}
				if len(msg.Topics) == 0 {
					continue
				}
				res := wp.get()
				// get connections for the particular topic
				wp.storage.GetByPtr(msg.Topics, res)
				if len(res) == 0 {
					wp.log.Info("no such topic", "topic", msg.Topics)
					wp.put(res)
					continue
				}

				for i := range res {
					c, ok := wp.connections.Load(i)
					if !ok {
						wp.log.Warn("the user disconnected connection before the message being written to it", "broker", msg.Broker, "topics", msg.Topics)
						continue
					}

					conn := c.(*connection.Connection)
					err := conn.Write(websocket.BinaryMessage, msg.Payload)
					if err != nil {
						wp.log.Error("error sending payload over the connection", "broker", msg.Broker, "topics", msg.Topics)
						wp.put(res)
						continue
					}
				}

				wp.put(res)
			case <-wp.exit:
				wp.log.Info("get exit signal, exiting from the workers pool")
				return
			}
		}
	}()
}
