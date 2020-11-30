package tests

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	j "github.com/json-iterator/go"
	"github.com/spiral/endure"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/gzip"
	httpPlugin "github.com/spiral/roadrunner/v2/plugins/http"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/spiral/roadrunner/v2/plugins/static"
	"github.com/stretchr/testify/assert"
)

var json = j.ConfigCompatibleWithStandardLibrary

func TestStaticPlugin(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel), endure.Visualize(endure.StdOut, ""))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http-static.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Gzip{},
		&static.Plugin{},
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

	t.Run("ServeSample", serveStaticSample)
	wg.Wait()
}

func serveStaticSample(t *testing.T) {
	b, r, err := get("http://localhost:21603/sample.txt")
	assert.NoError(t, err)
	assert.Equal(t, "sample", b)
	_ = r.Body.Close()
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

	err = r.Body.Close()
	if err != nil {
		return "", nil, err
	}

	return string(b), r, err
}

func tmpDir() string {
	p := os.TempDir()

	r, _ := j.Marshal(p)

	return string(r)
}

func all(fn string) string {
	f, _ := os.Open(fn)

	b := new(bytes.Buffer)
	_, err := io.Copy(b, f)
	if err != nil {
		return ""
	}

	err = f.Close()
	if err != nil {
		return ""
	}

	return b.String()
}
