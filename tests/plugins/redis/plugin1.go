package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spiral/errors"
	redisPlugin "github.com/spiral/roadrunner/v2/plugins/redis"
)

type Plugin1 struct {
	redisClient redis.UniversalClient
}

func (p *Plugin1) Init(redis redisPlugin.Redis) error {
	var err error
	p.redisClient, err = redis.RedisClient("redis")

	return err
}

func (p *Plugin1) Serve() chan error {
	const op = errors.Op("plugin1 serve")
	errCh := make(chan error, 1)
	p.redisClient.Set(context.Background(), "foo", "bar", time.Minute)

	stringCmd := p.redisClient.Get(context.Background(), "foo")
	data, err := stringCmd.Result()
	if err != nil {
		errCh <- errors.E(op, err)
		return errCh
	}

	if data != "bar" {
		errCh <- errors.E(op, errors.Str("no such key"))
		return errCh
	}

	return errCh
}

func (p *Plugin1) Stop() error {
	return nil
}
