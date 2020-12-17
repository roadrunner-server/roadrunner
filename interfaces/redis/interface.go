package redis

import "github.com/go-redis/redis/v8"

type Redis interface {
	GetClient() *redis.Client
	GetUniversalClient() *redis.UniversalClient
	GetClusterClient() *redis.ClusterClient
	GetSentinelClient() *redis.SentinelClient
}
