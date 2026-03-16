package tests

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	mocklogger "tests/mock"
	"tests/proto/service"

	"github.com/roadrunner-server/config/v5"
	"github.com/roadrunner-server/endure/v2"
	grpcPlugin "github.com/roadrunner-server/grpc/v5"
	rpcPlugin "github.com/roadrunner-server/rpc/v5"
	"github.com/roadrunner-server/server/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestGrpcPing verifies the full gRPC lifecycle: container startup, PHP worker
// registration of the Echo service, sending a Ping RPC, and validating the
// response (PHP uppercases the message).
func TestGrpcPing(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2024.1.0",
		Path:    "configs/.rr-grpc.yaml",
	}

	l, _ := mocklogger.ZapTestLogger(zap.DebugLevel)

	err := cont.RegisterAll(
		cfg,
		&grpcPlugin.Plugin{},
		&rpcPlugin.Plugin{},
		&server.Plugin{},
		l,
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

	stopCh := make(chan struct{}, 1)

	wg := &sync.WaitGroup{}
	wg.Go(func() {
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				stopErr := cont.Stop()
				if stopErr != nil {
					assert.FailNow(t, "error", stopErr.Error())
				}
			case <-sig:
				stopErr := cont.Stop()
				if stopErr != nil {
					assert.FailNow(t, "error", stopErr.Error())
				}
				return
			case <-stopCh:
				stopErr := cont.Stop()
				if stopErr != nil {
					assert.FailNow(t, "error", stopErr.Error())
				}
				return
			}
		}
	})

	time.Sleep(time.Second)

	t.Run("PingEcho", func(t *testing.T) {
		conn, errDial := grpc.NewClient(
			"127.0.0.1:9191",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		require.NoError(t, errDial)
		require.NotNil(t, conn)
		defer func() { _ = conn.Close() }()

		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()

		client := service.NewEchoClient(conn)
		resp, errPing := client.Ping(ctx, &service.Message{Msg: "hello"})
		require.NoError(t, errPing)
		require.Equal(t, "HELLO", resp.GetMsg())
	})

	stopCh <- struct{}{}
	wg.Wait()
}
