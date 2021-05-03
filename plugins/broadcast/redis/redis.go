package redis

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/go-redis/redis/v8"
)

// Redis based broadcast Router.
type Redis struct {
	client        redis.UniversalClient
	psClient      redis.UniversalClient
	router        *Router
	messages      chan *Message
	listen, leave chan subscriber
	stop          chan interface{}
	stopped       int32
}

// creates new redis broker
func redisBroker(cfg *RedisConfig) (*Redis, error) {
	client := cfg.redisClient()
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	psClient := cfg.redisClient()
	if _, err := psClient.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return &Redis{
		client:   client,
		psClient: psClient,
		router:   NewRouter(),
		messages: make(chan *Message),
		listen:   make(chan subscriber),
		leave:    make(chan subscriber),
		stop:     make(chan interface{}),
		stopped:  0,
	}, nil
}

// Serve serves broker.
func (r *Redis) Serve() error {
	pubsub := r.psClient.Subscribe(context.Background())
	channel := pubsub.Channel()

	for {
		select {
		case ctx := <-r.listen:
			ctx.done <- r.handleJoin(ctx, pubsub)
		case ctx := <-r.leave:
			ctx.done <- r.handleLeave(ctx, pubsub)
		case msg := <-channel:
			r.router.Dispatch(&Message{
				Topic:   msg.Channel,
				Payload: []byte(msg.Payload),
			})
		case <-r.stop:
			return nil
		}
	}
}

func (r *Redis) handleJoin(sub subscriber, pubsub *redis.PubSub) error {
	if sub.pattern != "" {
		newPatterns, err := r.router.SubscribePattern(sub.upstream, sub.pattern)
		if err != nil || len(newPatterns) == 0 {
			return err
		}

		return pubsub.PSubscribe(context.Background(), newPatterns...)
	}

	newTopics := r.router.Subscribe(sub.upstream, sub.topics...)
	if len(newTopics) == 0 {
		return nil
	}

	return pubsub.Subscribe(context.Background(), newTopics...)
}

func (r *Redis) handleLeave(sub subscriber, pubsub *redis.PubSub) error {
	if sub.pattern != "" {
		dropPatterns := r.router.UnsubscribePattern(sub.upstream, sub.pattern)
		if len(dropPatterns) == 0 {
			return nil
		}

		return pubsub.PUnsubscribe(context.Background(), dropPatterns...)
	}

	dropTopics := r.router.Unsubscribe(sub.upstream, sub.topics...)
	if len(dropTopics) == 0 {
		return nil
	}

	return pubsub.Unsubscribe(context.Background(), dropTopics...)
}

// Stop closes the consumption and disconnects broker.
func (r *Redis) Stop() {
	if atomic.CompareAndSwapInt32(&r.stopped, 0, 1) {
		close(r.stop)
	}
}

// Subscribe broker to one or multiple channels.
func (r *Redis) Subscribe(upstream chan *Message, topics ...string) error {
	if atomic.LoadInt32(&r.stopped) == 1 {
		return errors.New("broker has been stopped")
	}

	ctx := subscriber{upstream: upstream, topics: topics, done: make(chan error)}

	r.listen <- ctx
	return <-ctx.done
}

// SubscribePattern broker to pattern.
func (r *Redis) SubscribePattern(upstream chan *Message, pattern string) error {
	if atomic.LoadInt32(&r.stopped) == 1 {
		return errors.New("broker has been stopped")
	}

	ctx := subscriber{upstream: upstream, pattern: pattern, done: make(chan error)}

	r.listen <- ctx
	return <-ctx.done
}

// Unsubscribe broker from one or multiple channels.
func (r *Redis) Unsubscribe(upstream chan *Message, topics ...string) error {
	if atomic.LoadInt32(&r.stopped) == 1 {
		return errors.New("broker has been stopped")
	}

	ctx := subscriber{upstream: upstream, topics: topics, done: make(chan error)}

	r.leave <- ctx
	return <-ctx.done
}

// UnsubscribePattern broker from pattern.
func (r *Redis) UnsubscribePattern(upstream chan *Message, pattern string) error {
	if atomic.LoadInt32(&r.stopped) == 1 {
		return errors.New("broker has been stopped")
	}

	ctx := subscriber{upstream: upstream, pattern: pattern, done: make(chan error)}

	r.leave <- ctx
	return <-ctx.done
}

// Publish one or multiple Channel.
func (r *Redis) Publish(messages ...*Message) error {
	if atomic.LoadInt32(&r.stopped) == 1 {
		return errors.New("broker has been stopped")
	}

	for _, msg := range messages {
		if err := r.client.Publish(context.Background(), msg.Topic, []byte(msg.Payload)).Err(); err != nil {
			return err
		}
	}

	return nil
}
