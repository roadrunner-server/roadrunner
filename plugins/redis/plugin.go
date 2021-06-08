package redis

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/spiral/errors"
	websocketsv1 "github.com/spiral/roadrunner/v2/pkg/proto/websockets/v1beta"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"google.golang.org/protobuf/proto"
)

const PluginName = "redis"

type Plugin struct {
	sync.RWMutex
	// config for RR integration
	cfg *Config
	// logger
	log logger.Logger
	// redis universal client
	universalClient redis.UniversalClient

	// fanIn implementation used to deliver messages from all channels to the single websocket point
	fanin *FanIn
}

func (p *Plugin) GetClient() redis.UniversalClient {
	return p.universalClient
}

func (p *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("redis_plugin_init")

	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &p.cfg)
	if err != nil {
		return errors.E(op, errors.Disabled, err)
	}

	p.cfg.InitDefaults()
	p.log = log

	p.universalClient = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:              p.cfg.Addrs,
		DB:                 p.cfg.DB,
		Username:           p.cfg.Username,
		Password:           p.cfg.Password,
		SentinelPassword:   p.cfg.SentinelPassword,
		MaxRetries:         p.cfg.MaxRetries,
		MinRetryBackoff:    p.cfg.MaxRetryBackoff,
		MaxRetryBackoff:    p.cfg.MaxRetryBackoff,
		DialTimeout:        p.cfg.DialTimeout,
		ReadTimeout:        p.cfg.ReadTimeout,
		WriteTimeout:       p.cfg.WriteTimeout,
		PoolSize:           p.cfg.PoolSize,
		MinIdleConns:       p.cfg.MinIdleConns,
		MaxConnAge:         p.cfg.MaxConnAge,
		PoolTimeout:        p.cfg.PoolTimeout,
		IdleTimeout:        p.cfg.IdleTimeout,
		IdleCheckFrequency: p.cfg.IdleCheckFreq,
		ReadOnly:           p.cfg.ReadOnly,
		RouteByLatency:     p.cfg.RouteByLatency,
		RouteRandomly:      p.cfg.RouteRandomly,
		MasterName:         p.cfg.MasterName,
	})

	// init fanin
	p.fanin = newFanIn(p.universalClient, log)

	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (p *Plugin) Stop() error {
	const op = errors.Op("redis_plugin_stop")
	err := p.fanin.stop()
	if err != nil {
		return errors.E(op, err)
	}

	err = p.universalClient.Close()
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (p *Plugin) Name() string {
	return PluginName
}

// Available interface implementation
func (p *Plugin) Available() {}

func (p *Plugin) Publish(msg []byte) error {
	p.Lock()
	defer p.Unlock()

	m := &websocketsv1.Message{}
	err := proto.Unmarshal(msg, m)
	if err != nil {
		return errors.E(err)
	}

	for j := 0; j < len(m.GetTopics()); j++ {
		f := p.universalClient.Publish(context.Background(), m.GetTopics()[j], msg)
		if f.Err() != nil {
			return f.Err()
		}
	}
	return nil
}

func (p *Plugin) PublishAsync(msg []byte) {
	go func() {
		p.Lock()
		defer p.Unlock()
		m := &websocketsv1.Message{}
		err := proto.Unmarshal(msg, m)
		if err != nil {
			p.log.Error("message unmarshal error")
			return
		}

		for j := 0; j < len(m.GetTopics()); j++ {
			f := p.universalClient.Publish(context.Background(), m.GetTopics()[j], msg)
			if f.Err() != nil {
				p.log.Error("redis publish", "error", f.Err())
			}
		}
	}()
}

func (p *Plugin) Subscribe(connectionID string, topics ...string) error {
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
			p.log.Warn("could not subscribe to the provided topic", "connectionID", connectionID, "topic", topics[i])
			continue
		}
	}

	// and subscribe after
	return p.fanin.sub(topics...)
}

func (p *Plugin) Unsubscribe(connectionID string, topics ...string) error {
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
		err = p.fanin.unsub(topics[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Plugin) Connections(topic string, res map[string]struct{}) {
	hget := p.universalClient.SMembersMap(context.Background(), topic)
	r, err := hget.Result()
	if err != nil {
		panic(err)
	}

	// assighn connections
	// res expected to be from the sync.Pool
	for k := range r {
		res[k] = struct{}{}
	}
}

// Next return next message
func (p *Plugin) Next() (*websocketsv1.Message, error) {
	return <-p.fanin.consume(), nil
}
