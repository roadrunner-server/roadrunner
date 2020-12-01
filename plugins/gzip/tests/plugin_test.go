package tests

import (
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/spiral/endure"
	"github.com/spiral/roadrunner/v2/mocks"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/gzip"
	httpPlugin "github.com/spiral/roadrunner/v2/plugins/http"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/stretchr/testify/assert"
)

func TestGzipPlugin(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http-withGzip.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Gzip{},
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

	go func() {
		tt := time.NewTimer(time.Second * 10)
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
			case <-tt.C:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	t.Run("GzipCheckHeader", headerCheck)
	wg.Wait()
}

func headerCheck(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:18953", nil)
	assert.NoError(t, err)
	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: false,
		},
	}

	r, err := client.Do(req)
	assert.NoError(t, err)
	assert.True(t, r.Uncompressed)

	err = r.Body.Close()
	assert.NoError(t, err)
}

func TestMiddlewareNotExist(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http-middlewareNotExist.yaml",
		Prefix: "rr",
	}

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	mockLogger.EXPECT().Warn("requested middleware does not exist", "requested", "foo")

	err = cont.RegisterAll(
		cfg,
		mockLogger,
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Gzip{},
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

	go func() {
		tt := time.NewTimer(time.Second * 5)
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
			case <-tt.C:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	wg.Wait()
}
