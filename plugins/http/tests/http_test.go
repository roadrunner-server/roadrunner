package tests

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/spiral/endure"
	"github.com/spiral/roadrunner/v2/plugins/config"
	httpPlugin "github.com/spiral/roadrunner/v2/plugins/http"
	"github.com/spiral/roadrunner/v2/plugins/logger"

	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/stretchr/testify/assert"
)

func TestHTTPInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel), endure.Visualize(endure.StdOut, ""))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
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

	tt := time.NewTimer(time.Second * 10)
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

func TestHTTPHandler(t *testing.T) {
	//cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel), endure.Visualize(endure.StdOut, ""))
	//assert.NoError(t, err)
	//
	//cfg := &config.Viper{
	//	Path:   "configs/.rr-handler-echo.yaml",
	//	Prefix: "rr",
	//}
	//
	//err = cont.RegisterAll(
	//	cfg,
	//	&rpcPlugin.Plugin{},
	//	&logger.ZapLogger{},
	//	&server.Plugin{},
	//	&httpPlugin.Plugin{},
	//)
	//assert.NoError(t, err)
	//
	//err = cont.Init()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//ch, err := cont.Serve()
	//assert.NoError(t, err)
	//
	//sig := make(chan os.Signal, 1)
	//signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	//
	//go func() {
	//	tt := time.NewTimer(time.Minute * 5)
	//	for {
	//		select {
	//		case e := <-ch:
	//			assert.Fail(t, "error", e.Error.Error())
	//			err = cont.Stop()
	//			if err != nil {
	//				assert.FailNow(t, "error", err.Error())
	//			}
	//		case <-sig:
	//			err = cont.Stop()
	//			if err != nil {
	//				assert.FailNow(t, "error", err.Error())
	//			}
	//			return
	//		case <-tt.C:
	//			// timeout
	//			err = cont.Stop()
	//			if err != nil {
	//				assert.FailNow(t, "error", err.Error())
	//			}
	//			return
	//		}
	//	}
	//}()
}

func get(url string) (string, *http.Response, error) {
	r, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", nil, err
	}
	defer func() {
		_ = r.Body.Close()
	}()
	return string(b), r, err
}

// get request and return body
func getHeader(url string, h map[string]string) (string, *http.Response, error) {
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
	if err != nil {
		return "", nil, err
	}

	for k, v := range h {
		req.Header.Set(k, v)
	}

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", nil, err
	}

	err = r.Body.Close()
	if err != nil {
		return "", nil, err
	}
	return string(b), r, err
}
