package redis

import (
	"github.com/go-redis/redis/v8"
	"github.com/spiral/roadrunner/v2/interfaces/config"
	"github.com/spiral/roadrunner/v2/interfaces/log"
)

const PluginName = "redis"

type Plugin struct {
	// config for RR integration
	cfg *Config
	// redis client
	universalClient *redis.UniversalClient
	clusterClient   *redis.ClusterClient
	client          *redis.Client
	sentinelClient  *redis.SentinelClient
}

func (s *Plugin) GetClient() *redis.Client {
	return s.client
}

func (s *Plugin) GetUniversalClient() *redis.UniversalClient {
	return s.universalClient
}

func (s *Plugin) GetClusterClient() *redis.ClusterClient {
	return s.clusterClient
}

func (s *Plugin) GetSentinelClient() *redis.SentinelClient {
	return s.sentinelClient
}

func (s *Plugin) Init(cfg config.Configurer, log log.Logger) error {
	_ = cfg
	_ = log
	_ = s.cfg
	return nil
}

func (s *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	return errCh
}

func (s Plugin) Stop() error {
	return nil
}

func (s *Plugin) Name() string {
	return PluginName
}
