package jobs

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	toxiproxy "github.com/Shopify/toxiproxy/client"
	"github.com/golang/mock/gomock"
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner/v2/plugins/amqp"
	"github.com/spiral/roadrunner/v2/plugins/beanstalk"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/informer"
	"github.com/spiral/roadrunner/v2/plugins/jobs"
	"github.com/spiral/roadrunner/v2/plugins/resetter"
	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/spiral/roadrunner/v2/plugins/sqs"
	"github.com/spiral/roadrunner/v2/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDurabilityAMQP(t *testing.T) {
	client := toxiproxy.NewClient("127.0.0.1:8474")

	_, err := client.CreateProxy("redial", "127.0.0.1:23679", "127.0.0.1:5672")
	require.NoError(t, err)
	defer deleteProxy("redial", t)

	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	require.NoError(t, err)

	cfg := &config.Viper{
		Path:   "durability/.rr-amqp-durability-redial.yaml",
		Prefix: "rr",
	}

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	// general
	mockLogger.EXPECT().Debug("worker destructed", "pid", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("worker constructed", "pid", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("Started RPC service", "address", "tcp://127.0.0.1:6001", "plugins", gomock.Any()).Times(1)

	mockLogger.EXPECT().Info("pipeline active", "pipeline", "test-1", "start", gomock.Any(), "elapsed", gomock.Any()).Times(2)
	mockLogger.EXPECT().Info("pipeline active", "pipeline", "test-2", "start", gomock.Any(), "elapsed", gomock.Any()).Times(2)

	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogger.EXPECT().Warn("pipeline stopped", "pipeline", "test-1", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)
	mockLogger.EXPECT().Warn("pipeline stopped", "pipeline", "test-2", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)

	mockLogger.EXPECT().Info("job pushed to the queue", "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("job processing started", "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("job processed without errors", "ID", gomock.Any(), "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Error("job push error, job might be lost", "error", gomock.Any(), "pipeline", "test-1", "ID", gomock.Any(), "driver", "amqp", "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Error("job push error, job might be lost", "error", gomock.Any(), "pipeline", "test-2", "ID", gomock.Any(), "driver", "amqp", "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)

	mockLogger.EXPECT().Info("delivery channel closed, leaving the rabbit listener").Times(4)

	// redial errors
	mockLogger.EXPECT().Warn("rabbitmq reconnecting, caused by", "error", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error("pipeline error", "pipeline", "test-1", "error", gomock.Any(), "start", gomock.Any(), "elapsed", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error("pipeline error", "pipeline", "test-2", "error", gomock.Any(), "start", gomock.Any(), "elapsed", gomock.Any()).AnyTimes()

	mockLogger.EXPECT().Info("rabbitmq dial succeed. trying to redeclare queues and subscribers").AnyTimes()
	mockLogger.EXPECT().Info("queues and subscribers redeclared successfully").AnyTimes()

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		mockLogger,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&amqp.Plugin{},
	)
	require.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

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

	time.Sleep(time.Second * 3)
	disableProxy("redial", t)
	time.Sleep(time.Second * 3)

	go func() {
		time.Sleep(time.Second * 5)
		enableProxy("redial", t)
	}()

	t.Run("PushPipelineWhileRedialing-1", pushToPipeErr("test-1"))
	t.Run("PushPipelineWhileRedialing-2", pushToPipeErr("test-2"))

	time.Sleep(time.Second * 15)
	t.Run("PushPipelineWhileRedialing-1", pushToPipe("test-1"))
	t.Run("PushPipelineWhileRedialing-2", pushToPipe("test-2"))

	time.Sleep(time.Second * 5)

	stopCh <- struct{}{}
	wg.Wait()
}

func TestDurabilitySQS(t *testing.T) {
	client := toxiproxy.NewClient("127.0.0.1:8474")

	_, err := client.CreateProxy("redial", "127.0.0.1:19324", "127.0.0.1:9324")
	require.NoError(t, err)
	defer deleteProxy("redial", t)

	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	require.NoError(t, err)

	cfg := &config.Viper{
		Path:   "durability/.rr-sqs-durability-redial.yaml",
		Prefix: "rr",
	}

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	// general
	mockLogger.EXPECT().Debug("worker destructed", "pid", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("worker constructed", "pid", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("Started RPC service", "address", "tcp://127.0.0.1:6001", "plugins", gomock.Any()).Times(1)

	mockLogger.EXPECT().Info("pipeline active", "pipeline", "test-1", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)
	mockLogger.EXPECT().Info("pipeline active", "pipeline", "test-2", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)

	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogger.EXPECT().Warn("pipeline stopped", "pipeline", "test-1", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)
	mockLogger.EXPECT().Warn("pipeline stopped", "pipeline", "test-2", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)

	mockLogger.EXPECT().Info("job pushed to the queue", "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("job processing started", "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("job processed without errors", "ID", gomock.Any(), "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)
	// redial errors
	mockLogger.EXPECT().Error("pipeline error", "pipeline", "test-1", "error", gomock.Any(), "start", gomock.Any(), "elapsed", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error("pipeline error", "pipeline", "test-2", "error", gomock.Any(), "start", gomock.Any(), "elapsed", gomock.Any()).AnyTimes()

	// stop
	mockLogger.EXPECT().Warn("sqs listener stopped").AnyTimes()
	mockLogger.EXPECT().Info("------> job poller stopped <------").AnyTimes()

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		mockLogger,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&sqs.Plugin{},
	)
	require.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

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

	time.Sleep(time.Second * 3)
	disableProxy("redial", t)
	time.Sleep(time.Second * 3)

	go func() {
		time.Sleep(time.Second)
		t.Run("PushPipelineWhileRedialing-1", pushToPipe("test-1"))
		time.Sleep(time.Second)
		t.Run("PushPipelineWhileRedialing-2", pushToPipe("test-2"))
	}()

	time.Sleep(time.Second * 5)
	enableProxy("redial", t)

	t.Run("PushPipelineWhileRedialing-1", pushToPipe("test-1"))
	t.Run("PushPipelineWhileRedialing-2", pushToPipe("test-2"))

	time.Sleep(time.Second * 10)

	stopCh <- struct{}{}
	wg.Wait()
}

func TestDurabilityBeanstalk(t *testing.T) {
	client := toxiproxy.NewClient("127.0.0.1:8474")

	_, err := client.CreateProxy("redial", "127.0.0.1:11400", "127.0.0.1:11300")
	require.NoError(t, err)
	defer deleteProxy("redial", t)

	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	require.NoError(t, err)

	cfg := &config.Viper{
		Path:   "durability/.rr-beanstalk-durability-redial.yaml",
		Prefix: "rr",
	}

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	// general
	mockLogger.EXPECT().Debug("worker destructed", "pid", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("worker constructed", "pid", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("Started RPC service", "address", "tcp://127.0.0.1:6001", "plugins", gomock.Any()).Times(1)

	mockLogger.EXPECT().Info("pipeline active", "pipeline", "test-1", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)
	mockLogger.EXPECT().Info("pipeline active", "pipeline", "test-2", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)

	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogger.EXPECT().Info("job pushed to the queue", "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Warn("pipeline stopped", "pipeline", "test-1", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)
	mockLogger.EXPECT().Warn("pipeline stopped", "pipeline", "test-2", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)

	mockLogger.EXPECT().Info("job processing started", "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("job processed without errors", "ID", gomock.Any(), "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)
	// redial errors
	mockLogger.EXPECT().Info("beanstalk redial was successful").MinTimes(2)
	mockLogger.EXPECT().Error("pipeline error", "pipeline", "test-1", "error", gomock.Any(), "start", gomock.Any(), "elapsed", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error("pipeline error", "pipeline", "test-2", "error", gomock.Any(), "start", gomock.Any(), "elapsed", gomock.Any()).AnyTimes()

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		mockLogger,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&beanstalk.Plugin{},
	)
	require.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

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

	time.Sleep(time.Second * 3)
	disableProxy("redial", t)
	time.Sleep(time.Second * 3)

	go func() {
		time.Sleep(time.Second * 2)
		t.Run("PushPipelineWhileRedialing-1", pushToPipe("test-1"))
		t.Run("PushPipelineWhileRedialing-2", pushToPipe("test-2"))
	}()

	time.Sleep(time.Second * 5)
	enableProxy("redial", t)

	t.Run("PushPipelineWhileRedialing-1", pushToPipe("test-1"))
	t.Run("PushPipelineWhileRedialing-2", pushToPipe("test-2"))

	time.Sleep(time.Second * 10)

	stopCh <- struct{}{}
	wg.Wait()
}
