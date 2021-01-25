package logger

import (
	"os"
	"os/signal"
	"sync"
	"testing"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	// config plugin
	vp := &config.Viper{}
	vp.Path = ".rr.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&Plugin{})
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&logger.ZapLogger{})
	if err != nil {
		t.Fatal(err)
	}

	err = container.Init()
	if err != nil {
		t.Fatal(err)
	}

	errCh, err := container.Serve()
	if err != nil {
		t.Fatal(err)
	}

	// stop by CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	stopCh := make(chan struct{}, 1)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-errCh:
				assert.NoError(t, e.Error)
				assert.NoError(t, container.Stop())
				return
			case <-c:
				err = container.Stop()
				assert.NoError(t, err)
				return
			case <-stopCh:
				assert.NoError(t, container.Stop())
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}
