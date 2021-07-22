package jobs

import (
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	endure "github.com/spiral/endure/pkg/container"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/informer"
	"github.com/spiral/roadrunner/v2/plugins/jobs"
	"github.com/spiral/roadrunner/v2/plugins/jobs/drivers/ephemeral"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/resetter"
	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/spiral/roadrunner/v2/plugins/server"
	jobsv1beta "github.com/spiral/roadrunner/v2/proto/jobs/v1beta"
	"github.com/spiral/roadrunner/v2/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestEphemeralInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "ephemeral/.rr-ephemeral-init.yaml",
		Prefix: "rr",
	}

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	// general
	mockLogger.EXPECT().Debug("worker destructed", "pid", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("worker constructed", "pid", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("Started RPC service", "address", "tcp://127.0.0.1:6001", "services", gomock.Any()).Times(1)
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogger.EXPECT().Info("pipeline started", "pipeline", "test-1", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)
	mockLogger.EXPECT().Info("pipeline started", "pipeline", "test-2", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)

	mockLogger.EXPECT().Info("driver initialized", "driver", "amqp", "start", gomock.Any()).Times(2)

	mockLogger.EXPECT().Warn("pipeline stopped", "pipeline", "test-1", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)
	mockLogger.EXPECT().Warn("pipeline stopped", "pipeline", "test-2", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)

	mockLogger.EXPECT().Info("delivery channel closed, leaving the rabbit listener").Times(2)

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		// mockLogger,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&ephemeral.Plugin{},
	)
	assert.NoError(t, err)

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
	stopCh <- struct{}{}
	wg.Wait()
}

func TestEphemeralDeclare(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "ephemeral/.rr-ephemeral-declare.yaml",
		Prefix: "rr",
	}

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	// general
	mockLogger.EXPECT().Debug("worker destructed", "pid", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("worker constructed", "pid", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("Started RPC service", "address", "tcp://127.0.0.1:6001", "services", gomock.Any()).Times(1)
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogger.EXPECT().Info("pipeline started", "pipeline", "test-1", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)
	mockLogger.EXPECT().Info("pipeline started", "pipeline", "test-2", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)

	mockLogger.EXPECT().Info("driver initialized", "driver", "amqp", "start", gomock.Any()).Times(2)

	mockLogger.EXPECT().Warn("pipeline stopped", "pipeline", "test-1", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)
	mockLogger.EXPECT().Warn("pipeline stopped", "pipeline", "test-2", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)

	mockLogger.EXPECT().Info("delivery channel closed, leaving the rabbit listener").Times(2)

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		// mockLogger,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&ephemeral.Plugin{},
	)
	assert.NoError(t, err)

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

	t.Run("DeclareEphemeralPipeline", declareEphemeralPipe)
	t.Run("ConsumeEphemeralPipeline", consumeEphemeralPipe)
	t.Run("PushEphemeralPipeline", pushToPipe("test-3"))
	t.Run("PauseEphemeralPipeline", pausePipelines("test-3"))
	t.Run("DestroyEphemeralPipeline", destroyPipelines("test-3"))

	time.Sleep(time.Second * 5)
	stopCh <- struct{}{}
	wg.Wait()
}

func declareEphemeralPipe(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	pipe := &jobsv1beta.DeclareRequest{Pipeline: map[string]string{
		"driver":   "ephemeral",
		"name":     "test-3",
		"prefetch": "10000",
	}}

	er := &jobsv1beta.Empty{}
	err = client.Call("jobs.Declare", pipe, er)
	assert.NoError(t, err)
}

func consumeEphemeralPipe(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	pipe := &jobsv1beta.Pipelines{Pipelines: make([]string, 1)}
	pipe.GetPipelines()[0] = "test-3"

	er := &jobsv1beta.Empty{}
	err = client.Call("jobs.Resume", pipe, er)
	assert.NoError(t, err)
}
