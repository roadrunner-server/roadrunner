package redis

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type PubSubDriver struct {
	sync.RWMutex
	cfg *Config `mapstructure:"redis"`

	log             logger.Logger
	channel         *redisChannel
	universalClient redis.UniversalClient
	stopCh          chan struct{}
}

func NewPubSubDriver(log logger.Logger, key string, cfgPlugin config.Configurer, stopCh chan struct{}) (pubsub.PubSub, error) {
	const op = errors.Op("new_pub_sub_driver")
	ps := &PubSubDriver{
		log:    log,
		stopCh: stopCh,
	}

	// will be different for every connected driver
	err := cfgPlugin.UnmarshalKey(key, &ps.cfg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	ps.cfg.InitDefaults()

	ps.universalClient = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:              ps.cfg.Addrs,
		DB:                 ps.cfg.DB,
		Username:           ps.cfg.Username,
		Password:           ps.cfg.Password,
		SentinelPassword:   ps.cfg.SentinelPassword,
		MaxRetries:         ps.cfg.MaxRetries,
		MinRetryBackoff:    ps.cfg.MaxRetryBackoff,
		MaxRetryBackoff:    ps.cfg.MaxRetryBackoff,
		DialTimeout:        ps.cfg.DialTimeout,
		ReadTimeout:        ps.cfg.ReadTimeout,
		WriteTimeout:       ps.cfg.WriteTimeout,
		PoolSize:           ps.cfg.PoolSize,
		MinIdleConns:       ps.cfg.MinIdleConns,
		MaxConnAge:         ps.cfg.MaxConnAge,
		PoolTimeout:        ps.cfg.PoolTimeout,
		IdleTimeout:        ps.cfg.IdleTimeout,
		IdleCheckFrequency: ps.cfg.IdleCheckFreq,
		ReadOnly:           ps.cfg.ReadOnly,
		RouteByLatency:     ps.cfg.RouteByLatency,
		RouteRandomly:      ps.cfg.RouteRandomly,
		MasterName:         ps.cfg.MasterName,
	})

	statusCmd := ps.universalClient.Ping(context.Background())
	if statusCmd.Err() != nil {
		return nil, statusCmd.Err()
	}

	ps.channel = newRedisChannel(ps.universalClient, log)

	ps.stop()

	return ps, nil
}

func (p *PubSubDriver) stop() {
	go func() {
		for range p.stopCh {
			_ = p.channel.stop()
			return
		}
	}()
}

func (p *PubSubDriver) Publish(msg *pubsub.Message) error {
	p.Lock()
	defer p.Unlock()

	f := p.universalClient.Publish(context.Background(), msg.Topic, msg.Payload)
	if f.Err() != nil {
		return f.Err()
	}

	return nil
}

func (p *PubSubDriver) PublishAsync(msg *pubsub.Message) {
	go func() {
		p.Lock()
		defer p.Unlock()

		f := p.universalClient.Publish(context.Background(), msg.Topic, msg.Payload)
		if f.Err() != nil {
			p.log.Error("redis publish", "error", f.Err())
		}
	}()
}

func (p *PubSubDriver) Subscribe(connectionID string, topics ...string) error {
	// just add a connection
	for i := 0; i < len(topics); i++ {
		// key - topic
		// value - connectionID
		hset := p.universalClient.SAdd(context.Background(), topics[i], connectionID)
		res, err := hset.Result()
		if err != nil {
			return err
		}
		if res == 0 {
			p.log.Warn("could not subscribe to the provided topic, you might be already subscribed to it", "connectionID", connectionID, "topic", topics[i])
			continue
		}
	}

	// and subscribe after
	return p.channel.sub(topics...)
}

func (p *PubSubDriver) Unsubscribe(connectionID string, topics ...string) error {
	// Remove topics from the storage
	for i := 0; i < len(topics); i++ {
		srem := p.universalClient.SRem(context.Background(), topics[i], connectionID)
		if srem.Err() != nil {
			return srem.Err()
		}
	}

	for i := 0; i < len(topics); i++ {
		// if there are no such topics, we can safely unsubscribe from the redis
		exists := p.universalClient.Exists(context.Background(), topics[i])
		res, err := exists.Result()
		if err != nil {
			return err
		}

		// if we have associated connections - skip
		if res == 1 { // exists means that topic still exists and some other nodes may have connections associated with it
			continue
		}

		// else - unsubscribe
		err = p.channel.unsub(topics[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PubSubDriver) Connections(topic string, res map[string]struct{}) {
	hget := p.universalClient.SMembersMap(context.Background(), topic)
	r, err := hget.Result()
	if err != nil {
		panic(err)
	}

	// assign connections
	// res expected to be from the sync.Pool
	for k := range r {
		res[k] = struct{}{}
	}
}

// Next return next message
func (p *PubSubDriver) Next() (*pubsub.Message, error) {
	return p.channel.message(), nil
}
