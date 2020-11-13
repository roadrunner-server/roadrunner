package tests

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/spiral/endure"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/metrics"
	"github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/stretchr/testify/assert"
)

// get request and return body
func get(url string) (string, error) {
	r, err := http.Get(url)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	err = r.Body.Close()
	if err != nil {
		return "", err
	}
	// unsafe
	return string(b), err
}

func TestMetricsInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Viper{}
	cfg.Prefix = "rr"
	cfg.Path = ".rr-test.yaml"

	err = cont.Register(cfg)
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&metrics.Plugin{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&rpc.Plugin{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&logger.ZapLogger{})
	if err != nil {
		t.Fatal(err)
	}
	err = cont.Register(&Plugin1{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	tt := time.NewTimer(time.Second * 5)

	out, err := get("http://localhost:2112/metrics")
	assert.NoError(t, err)

	assert.Contains(t, out, "go_gc_duration_seconds")

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

func TestMetricsGaugeCollector(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel), endure.Visualize(endure.StdOut, ""))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Viper{}
	cfg.Prefix = "rr"
	cfg.Path = ".rr-test.yaml"

	err = cont.Register(cfg)
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&metrics.Plugin{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&rpc.Plugin{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&logger.ZapLogger{})
	if err != nil {
		t.Fatal(err)
	}
	err = cont.Register(&Plugin1{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	time.Sleep(time.Second)
	tt := time.NewTimer(time.Second * 5)

	out, err := get("http://localhost:2112/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, "my_gauge 100")

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
