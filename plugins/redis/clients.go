package redis

import (
	"github.com/go-redis/redis/v8"
	"github.com/spiral/errors"
)

// RedisClient return a client based on the provided section key
// key sample: kv.some-section.redis
// kv.redis
// redis (root)
func (p *Plugin) RedisClient(key string) (redis.UniversalClient, error) {
	const op = errors.Op("redis_get_client")

	if !p.cfgPlugin.Has(key) {
		return nil, errors.E(op, errors.Errorf("no such section: %s", key))
	}

	cfg := &Config{}

	err := p.cfgPlugin.UnmarshalKey(key, cfg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	cfg.InitDefaults()

	uc := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:              cfg.Addrs,
		DB:                 cfg.DB,
		Username:           cfg.Username,
		Password:           cfg.Password,
		SentinelPassword:   cfg.SentinelPassword,
		MaxRetries:         cfg.MaxRetries,
		MinRetryBackoff:    cfg.MaxRetryBackoff,
		MaxRetryBackoff:    cfg.MaxRetryBackoff,
		DialTimeout:        cfg.DialTimeout,
		ReadTimeout:        cfg.ReadTimeout,
		WriteTimeout:       cfg.WriteTimeout,
		PoolSize:           cfg.PoolSize,
		MinIdleConns:       cfg.MinIdleConns,
		MaxConnAge:         cfg.MaxConnAge,
		PoolTimeout:        cfg.PoolTimeout,
		IdleTimeout:        cfg.IdleTimeout,
		IdleCheckFrequency: cfg.IdleCheckFreq,
		ReadOnly:           cfg.ReadOnly,
		RouteByLatency:     cfg.RouteByLatency,
		RouteRandomly:      cfg.RouteRandomly,
		MasterName:         cfg.MasterName,
	})

	return uc, nil
}

func (p *Plugin) DefaultClient() redis.UniversalClient {
	cfg := &Config{}
	cfg.InitDefaults()

	uc := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:              cfg.Addrs,
		DB:                 cfg.DB,
		Username:           cfg.Username,
		Password:           cfg.Password,
		SentinelPassword:   cfg.SentinelPassword,
		MaxRetries:         cfg.MaxRetries,
		MinRetryBackoff:    cfg.MaxRetryBackoff,
		MaxRetryBackoff:    cfg.MaxRetryBackoff,
		DialTimeout:        cfg.DialTimeout,
		ReadTimeout:        cfg.ReadTimeout,
		WriteTimeout:       cfg.WriteTimeout,
		PoolSize:           cfg.PoolSize,
		MinIdleConns:       cfg.MinIdleConns,
		MaxConnAge:         cfg.MaxConnAge,
		PoolTimeout:        cfg.PoolTimeout,
		IdleTimeout:        cfg.IdleTimeout,
		IdleCheckFrequency: cfg.IdleCheckFreq,
		ReadOnly:           cfg.ReadOnly,
		RouteByLatency:     cfg.RouteByLatency,
		RouteRandomly:      cfg.RouteRandomly,
		MasterName:         cfg.MasterName,
	})

	return uc
}
