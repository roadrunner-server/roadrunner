package redis

import (
	"context"
	"sync"

	"github.com/spiral/roadrunner/v2/pkg/pubsub/message"
	"github.com/spiral/roadrunner/v2/plugins/logger"

	"github.com/go-redis/redis/v8"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/utils"
)

type FanIn struct {
	sync.Mutex

	client redis.UniversalClient
	pubsub *redis.PubSub

	log logger.Logger

	// out channel with all subs
	out chan *message.Message

	exit chan struct{}
}

func NewFanIn(redisClient redis.UniversalClient, log logger.Logger) *FanIn {
	out := make(chan *message.Message, 100)
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

func (fi *FanIn) AddChannel(topics ...string) error {
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
			fi.out <- message.GetRootAsMessage(utils.AsBytes(msg.Payload), 0)
		case <-fi.exit:
			return
		}
	}
}

func (fi *FanIn) RemoveChannel(topics ...string) error {
	const op = errors.Op("fanin_remove")
	err := fi.pubsub.Unsubscribe(context.Background(), topics...)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (fi *FanIn) Stop() error {
	fi.exit <- struct{}{}
	close(fi.out)
	close(fi.exit)
	return nil
}

func (fi *FanIn) Consume() <-chan *message.Message {
	return fi.out
}
