package debug_test

import (
	"context"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/spiral/roadrunner-binary/v2/internal/debug"

	"github.com/stretchr/testify/assert"
)

func TestServer_StartingAndStopping(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var (
		s    = debug.NewServer()
		port = strconv.Itoa(rand.Intn(10000) + 10000) //nolint:gosec
	)

	go func() { assert.ErrorIs(t, s.Start(":"+port), http.ErrServerClosed) }()

	defer func() { assert.NoError(t, s.Stop(context.Background())) }()

	for i := 0; i < 100; i++ { // wait for server started state
		if l, err := net.Dial("tcp", ":"+port); err != nil {
			<-time.After(time.Millisecond)
		} else {
			_ = l.Close()

			break
		}
	}

	for _, uri := range []string{ // assert that pprof handlers exists
		"http://127.0.0.1:" + port + "/debug/pprof/",
		"http://127.0.0.1:" + port + "/debug/pprof/cmdline",
		// "http://127.0.0.1:" + port + "/debug/pprof/profile",
		"http://127.0.0.1:" + port + "/debug/pprof/symbol",
		// "http://127.0.0.1:" + port + "/debug/pprof/trace",
	} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)

		req, _ := http.NewRequestWithContext(ctx, http.MethodHead, uri, http.NoBody)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		_ = resp.Body.Close()

		cancel()
	}
}
