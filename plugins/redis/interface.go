package redis

import "github.com/go-redis/redis/v8"

// Redis in the redis KV plugin interface
type Redis interface {
	// RedisClient provides universal redis client
	RedisClient(key string) (redis.UniversalClient, error)

	// DefaultClient provide default redis client based on redis defaults
	DefaultClient() redis.UniversalClient
}
