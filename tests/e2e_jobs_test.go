package tests

import (
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	mocklogger "tests/mock"

	"tests/helpers"

	"github.com/roadrunner-server/config/v5"
	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/informer/v5"
	"github.com/roadrunner-server/jobs/v5"
	"github.com/roadrunner-server/memory/v5"
	"github.com/roadrunner-server/resetter/v5"
	rpcPlugin "github.com/roadrunner-server/rpc/v5"
	"github.com/roadrunner-server/server/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestJobsInMemory verifies the full jobs lifecycle using the in-memory driver:
// container startup, pipeline consumption, job push, processing, and teardown.
func TestJobsInMemory(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2024.1.0",
		Path:    "configs/.rr-jobs-memory.yaml",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)

	err := cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&jobs.Plugin{},
		&memory.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
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

	time.Sleep(time.Second * 3)

	t.Run("PushToTest1", helpers.PushToPipe("test-1", false, "127.0.0.1:6201"))
	t.Run("PushToTest2", helpers.PushToPipe("test-2", false, "127.0.0.1:6201"))

	time.Sleep(time.Second * 2)

	t.Run("DestroyPipelines", helpers.DestroyPipelines("127.0.0.1:6201", "test-1", "test-2"))

	stopCh <- struct{}{}
	wg.Wait()

	require.GreaterOrEqual(t, oLogger.FilterMessageSnippet("pipeline was started").Len(), 2)
	require.GreaterOrEqual(t, oLogger.FilterMessageSnippet("pipeline was stopped").Len(), 2)
	require.GreaterOrEqual(t, oLogger.FilterMessageSnippet("job was pushed successfully").Len(), 2)
}
