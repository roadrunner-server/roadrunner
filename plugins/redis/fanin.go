package redis

import (
	"context"
	"sync"

	websocketsv1 "github.com/spiral/roadrunner/v2/pkg/proto/websockets/v1beta"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"google.golang.org/protobuf/proto"

	"github.com/go-redis/redis/v8"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/utils"
)

type FanIn struct {
	sync.Mutex

	// redis client
	client redis.UniversalClient
	pubsub *redis.PubSub

	log logger.Logger

	// out channel with all subs
	out chan *websocketsv1.Message

	exit chan struct{}
}

func newFanIn(redisClient redis.UniversalClient, log logger.Logger) *FanIn {
	out := make(chan *websocketsv1.Message, 100)
	fi := &FanIn{
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

func (fi *FanIn) sub(topics ...string) error {
	const op = errors.Op("fanin_addchannel")
	err := fi.pubsub.Subscribe(context.Background(), topics...)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// read reads messages from the pubsub subscription
func (fi *FanIn) read() {
	for {
		select {
		// here we receive message from us (which we sent before in Publish)
		// it should be compatible with the websockets.Msg interface
		// payload should be in the redis.message.payload field

		case msg, ok := <-fi.pubsub.Channel():
			// channel closed
			if !ok {
				return
			}

			m := &websocketsv1.Message{}
			err := proto.Unmarshal(utils.AsBytes(msg.Payload), m)
			if err != nil {
				fi.log.Error("message unmarshal")
				continue
			}

			fi.out <- m
		case <-fi.exit:
			return
		}
	}
}

func (fi *FanIn) unsub(topic string) error {
	const op = errors.Op("fanin_remove")
	err := fi.pubsub.Unsubscribe(context.Background(), topic)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (fi *FanIn) stop() error {
	fi.exit <- struct{}{}
	close(fi.out)
	close(fi.exit)
	return nil
}

func (fi *FanIn) consume() <-chan *websocketsv1.Message {
	return fi.out
}
