package broadcast

import (
	"errors"

	"github.com/go-redis/redis/v8"
)

// Config configures the broadcast extension.
type Config struct {
	// RedisConfig configures redis broker.
	Redis *RedisConfig
}

// Hydrate reads the configuration values from the source configuration.
//func (c *Config) Hydrate(cfg service.Config) error {
//	if err := cfg.Unmarshal(c); err != nil {
//		return err
//	}
//
//	if c.Redis != nil {
//		return c.Redis.isValid()
//	}
//
//	return nil
//}

// InitDefaults enables in memory broadcast configuration.
func (c *Config) InitDefaults() error {
	return nil
}

// RedisConfig configures redis broker.
type RedisConfig struct {
	// Addr of the redis server.
	Addr string

	// Password to redis server.
	Password string

	// DB index.
	DB int
}

// clusterOptions
func (cfg *RedisConfig) redisClient() redis.UniversalClient {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		PoolSize: 2,
	})
}

// check if redis config is valid.
func (cfg *RedisConfig) isValid() error {
	if cfg.Addr == "" {
		return errors.New("redis addr is required")
	}

	return nil
}
