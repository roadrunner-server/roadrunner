package tests

import (
	compressGzip "compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	mocklogger "tests/mock"

	"github.com/roadrunner-server/config/v5"
	"github.com/roadrunner-server/endure/v2"
	gzipPlugin "github.com/roadrunner-server/gzip/v5"
	"github.com/roadrunner-server/headers/v5"
	httpPlugin "github.com/roadrunner-server/http/v5"
	"github.com/roadrunner-server/prometheus/v5"
	proxyIP "github.com/roadrunner-server/proxy_ip_parser/v5"
	rpcPlugin "github.com/roadrunner-server/rpc/v5"
	"github.com/roadrunner-server/send/v5"
	"github.com/roadrunner-server/server/v5"
	"github.com/roadrunner-server/static/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestHTTPWithMiddleware verifies that the HTTP plugin works end-to-end with
// headers, gzip, prometheus metrics, proxy_ip_parser, and sendfile middleware
// all wired together via the Endure DI container.
func TestHTTPWithMiddleware(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2024.1.0",
		Path:    "configs/.rr-http-middleware.yaml",
	}

	l, _ := mocklogger.ZapTestLogger(zap.DebugLevel)

	err := cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&httpPlugin.Plugin{},
		&headers.Plugin{},
		&gzipPlugin.Plugin{},
		&prometheus.Plugin{},
		&proxyIP.Plugin{},
		&send.Plugin{},
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
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	})

	time.Sleep(time.Second)

	t.Run("EchoWithMiddleware", func(t *testing.T) {
		req, errReq := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://127.0.0.1:18950/?hello=world", nil)
		require.NoError(t, errReq)
		req.Header.Set("Accept-Encoding", "gzip")

		resp, errDo := http.DefaultClient.Do(req)
		require.NoError(t, errDo)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Verify the headers middleware added our custom response header.
		assert.Equal(t, "e2e-roadrunner", resp.Header.Get("X-Test"))

		// Verify gzip encoding is applied.
		assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

		// Decompress and verify the response body.
		gr, errGz := compressGzip.NewReader(resp.Body)
		require.NoError(t, errGz)
		defer func() { _ = gr.Close() }()

		body, errRead := io.ReadAll(gr)
		require.NoError(t, errRead)
		assert.Equal(t, "WORLD", string(body))
	})

	stopCh <- struct{}{}
	wg.Wait()
}

// TestHTTPStaticFile verifies that the static middleware serves files from disk,
// and that non-static requests fall through to the PHP worker.
func TestHTTPStaticFile(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2024.1.0",
		Path:    "configs/.rr-http-static.yaml",
	}

	l, _ := mocklogger.ZapTestLogger(zap.DebugLevel)

	err := cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&httpPlugin.Plugin{},
		&static.Plugin{},
		&gzipPlugin.Plugin{},
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
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	})

	time.Sleep(time.Second)

	t.Run("ServeStaticFile", func(t *testing.T) {
		req, errReq := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://127.0.0.1:18951/sample.txt", nil)
		require.NoError(t, errReq)

		resp, errDo := http.DefaultClient.Do(req)
		require.NoError(t, errDo)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, errRead := io.ReadAll(resp.Body)
		require.NoError(t, errRead)
		assert.Contains(t, string(body), "Hello from RoadRunner e2e static file test!")
	})

	t.Run("FallThroughToPHP", func(t *testing.T) {
		req, errReq := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://127.0.0.1:18951/?hello=world", nil)
		require.NoError(t, errReq)

		resp, errDo := http.DefaultClient.Do(req)
		require.NoError(t, errDo)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		body, errRead := io.ReadAll(resp.Body)
		require.NoError(t, errRead)
		assert.Equal(t, "WORLD", string(body))
	})

	stopCh <- struct{}{}
	wg.Wait()
}
