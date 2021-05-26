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

const PluginName = "redis"

type Plugin struct {
	sync.Mutex
	// config for RR integration
	cfg *Config
	// logger
	log logger.Logger
	// redis universal client
	universalClient redis.UniversalClient

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
	p.fanin = NewFanIn(p.universalClient, log)

	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (p *Plugin) Stop() error {
	const op = errors.Op("redis_plugin_stop")
	err := p.fanin.Stop()
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

func (p *Plugin) Publish(msg []pubsub.Message) error {
	p.Lock()
	defer p.Unlock()

	for i := 0; i < len(msg); i++ {
		for j := 0; j < len(msg[i].Topics()); j++ {
			f := p.universalClient.Publish(context.Background(), msg[i].Topics()[j], msg[i])
			if f.Err() != nil {
				return f.Err()
			}
		}
	}
	return nil
}

func (p *Plugin) PublishAsync(msg []pubsub.Message) {
	go func() {
		p.Lock()
		defer p.Unlock()
		for i := 0; i < len(msg); i++ {
			for j := 0; j < len(msg[i].Topics()); j++ {
				f := p.universalClient.Publish(context.Background(), msg[i].Topics()[j], msg[i])
				if f.Err() != nil {
					p.log.Error("errors publishing message", "topic", msg[i].Topics()[j], "error", f.Err().Error())
					continue
				}
			}
		}
	}()
}

func (p *Plugin) Subscribe(topics ...string) error {
	return p.fanin.AddChannel(topics...)
}

func (p *Plugin) Unsubscribe(topics ...string) error {
	return p.fanin.RemoveChannel(topics...)
}

// Next return next message
func (p *Plugin) Next() (pubsub.Message, error) {
	return <-p.fanin.Consume(), nil
}
