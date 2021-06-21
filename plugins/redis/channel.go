package redis

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/utils"
)

type redisChannel struct {
	sync.Mutex

	// redis client
	client redis.UniversalClient
	pubsub *redis.PubSub

	log logger.Logger

	// out channel with all subs
	out chan *pubsub.Message

	exit chan struct{}
}

func newRedisChannel(redisClient redis.UniversalClient, log logger.Logger) *redisChannel {
	out := make(chan *pubsub.Message, 100)
	fi := &redisChannel{
		out:    out,
		client: redisClient,
		pubsub: redisClient.Subscribe(context.Background()),
		exit:   make(chan struct{}),
		log:    log,
	}

	// start reading messages
	go fi.read()

	return fi
}

func (r *redisChannel) sub(topics ...string) error {
	const op = errors.Op("redis_sub")
	err := r.pubsub.Subscribe(context.Background(), topics...)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// read reads messages from the pubsub subscription
func (r *redisChannel) read() {
	for {
		select {
		// here we receive message from us (which we sent before in Publish)
		// it should be compatible with the pubsub.Message structure
		// payload should be in the redis.message.payload field

		case msg, ok := <-r.pubsub.Channel():
			// channel closed
			if !ok {
				return
			}

			r.out <- &pubsub.Message{
				Topic:   msg.Channel,
				Payload: utils.AsBytes(msg.Payload),
			}

		case <-r.exit:
			return
		}
	}
}

func (r *redisChannel) unsub(topic string) error {
	const op = errors.Op("redis_unsub")
	err := r.pubsub.Unsubscribe(context.Background(), topic)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (r *redisChannel) stop() error {
	r.exit <- struct{}{}
	close(r.out)
	close(r.exit)
	return nil
}

func (r *redisChannel) message() *pubsub.Message {
	return <-r.out
}
