package redis

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang/mock/gomock"
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/redis"
	"github.com/spiral/roadrunner/v2/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func redisConfig(port string) string {
	cfg := `
redis:
  addrs:
    - 'localhost:%s'
  master_name: ''
  username: ''
  password: ''
  db: 0
  sentinel_password: ''
  route_by_latency: false
  route_randomly: false
  dial_timeout: 0
  max_retries: 1
  min_retry_backoff: 0
  max_retry_backoff: 0
  pool_size: 0
  min_idle_conns: 0
  max_conn_age: 0
  read_timeout: 0
  write_timeout: 0
  pool_timeout: 0
  idle_timeout: 0
  idle_check_freq: 0
  read_only: false
`
	return fmt.Sprintf(cfg, port)
}

func TestRedisInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}

	s, err := miniredis.Run()
	assert.NoError(t, err)

	c := redisConfig(s.Port())

	cfg := &config.Viper{}
	cfg.Type = "yaml"
	cfg.ReadInCfg = []byte(c)

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	err = cont.RegisterAll(
		cfg,
		mockLogger,
		&redis.Plugin{},
		&Plugin1{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}
