package redis

import (
	"github.com/go-redis/redis/v8"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const PluginName = "redis"

type Plugin struct {
	// config for RR integration
	cfg *Config
	// logger
	log logger.Logger
	// redis universal client
	universalClient redis.UniversalClient
}

func (s *Plugin) GetClient() redis.UniversalClient {
	return s.universalClient
}

func (s *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("redis_plugin_init")

	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &s.cfg)
	if err != nil {
		return errors.E(op, errors.Disabled, err)
	}

	s.cfg.InitDefaults()
	s.log = log

	s.universalClient = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:              s.cfg.Addrs,
		DB:                 s.cfg.DB,
		Username:           s.cfg.Username,
		Password:           s.cfg.Password,
		SentinelPassword:   s.cfg.SentinelPassword,
		MaxRetries:         s.cfg.MaxRetries,
		MinRetryBackoff:    s.cfg.MaxRetryBackoff,
		MaxRetryBackoff:    s.cfg.MaxRetryBackoff,
		DialTimeout:        s.cfg.DialTimeout,
		ReadTimeout:        s.cfg.ReadTimeout,
		WriteTimeout:       s.cfg.WriteTimeout,
		PoolSize:           s.cfg.PoolSize,
		MinIdleConns:       s.cfg.MinIdleConns,
		MaxConnAge:         s.cfg.MaxConnAge,
		PoolTimeout:        s.cfg.PoolTimeout,
		IdleTimeout:        s.cfg.IdleTimeout,
		IdleCheckFrequency: s.cfg.IdleCheckFreq,
		ReadOnly:           s.cfg.ReadOnly,
		RouteByLatency:     s.cfg.RouteByLatency,
		RouteRandomly:      s.cfg.RouteRandomly,
		MasterName:         s.cfg.MasterName,
	})

	return nil
}

func (s *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s Plugin) Stop() error {
	return s.universalClient.Close()
}

func (s *Plugin) Name() string {
	return PluginName
}
