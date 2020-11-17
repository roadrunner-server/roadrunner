package test

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/spiral/endure"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/http"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	//rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/stretchr/testify/assert"
)

func TestHTTPInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel), endure.Visualize(endure.StdOut, ""))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   ".rr-http.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		//&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&http.Plugin{},
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

	tt := time.NewTimer(time.Minute * 3)
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
}
