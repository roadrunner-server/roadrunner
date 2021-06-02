package pool

import (
	"sync"

	"github.com/fasthttp/websocket"
	"github.com/spiral/roadrunner/v2/pkg/pubsub/message"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/websockets/connection"
	"github.com/spiral/roadrunner/v2/plugins/websockets/storage"
	"github.com/spiral/roadrunner/v2/utils"
)

type WorkersPool struct {
	storage     *storage.Storage
	connections *sync.Map
	resPool     sync.Pool
	log         logger.Logger

	queue chan *message.Message
	exit  chan struct{}
}

// NewWorkersPool constructs worker pool for the websocket connections
func NewWorkersPool(storage *storage.Storage, connections *sync.Map, log logger.Logger) *WorkersPool {
	wp := &WorkersPool{
		connections: connections,
		queue:       make(chan *message.Message, 100),
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

func (wp *WorkersPool) Queue(msg *message.Message) {
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
				_ = msg
				if msg == nil {
					continue
				}
				if msg.TopicsLength() == 0 {
					continue
				}
				res := wp.get()
				for i := 0; i < msg.TopicsLength(); i++ {
					// get connections for the particular topic
					wp.storage.GetOneByPtr(utils.AsString(msg.Topics(i)), res)
				}
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
					// TODO sync pool for the bytes
					bb := make([]byte, msg.PayloadLength())
					for i := 0; i < msg.PayloadLength(); i++ {
						bb[i] = byte(msg.Payload(i))
					}
					err := conn.Write(websocket.BinaryMessage, bb)
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
