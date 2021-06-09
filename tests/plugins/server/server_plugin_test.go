package server

import (
	"os"
	"os/signal"
	"sync"
	"testing"
	"time"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/stretchr/testify/assert"
)

func TestAppPipes(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"

	err = container.RegisterAll(
		vp,
		&server.Plugin{},
		&Foo{},
		&logger.ZapLogger{},
	)
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

	tt := time.NewTimer(time.Second * 10)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer tt.Stop()
		for {
			select {
			case e := <-errCh:
				assert.NoError(t, e.Error)
				assert.NoError(t, container.Stop())
				return
			case <-c:
				er := container.Stop()
				assert.NoError(t, er)
				return
			case <-tt.C:
				assert.NoError(t, container.Stop())
				return
			}
		}
	}()

	wg.Wait()
}

func TestAppSockets(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr-sockets.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&server.Plugin{})
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&Foo2{})
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

	// stop after 10 seconds
	tt := time.NewTicker(time.Second * 10)

	for {
		select {
		case e := <-errCh:
			assert.NoError(t, e.Error)
			assert.NoError(t, container.Stop())
			return
		case <-c:
			er := container.Stop()
			if er != nil {
				panic(er)
			}
			return
		case <-tt.C:
			tt.Stop()
			assert.NoError(t, container.Stop())
			return
		}
	}
}

func TestAppTCP(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr-tcp.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&server.Plugin{})
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&Foo3{})
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

	// stop after 10 seconds
	tt := time.NewTicker(time.Second * 10)

	for {
		select {
		case e := <-errCh:
			assert.NoError(t, e.Error)
			assert.NoError(t, container.Stop())
			return
		case <-c:
			er := container.Stop()
			if er != nil {
				panic(er)
			}
			return
		case <-tt.C:
			tt.Stop()
			assert.NoError(t, container.Stop())
			return
		}
	}
}

func TestAppWrongConfig(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rrrrrrrrrr.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&server.Plugin{})
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&Foo3{})
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&logger.ZapLogger{})
	if err != nil {
		t.Fatal(err)
	}

	assert.Error(t, container.Init())
}

func TestAppWrongRelay(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr-wrong-relay.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&server.Plugin{})
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&Foo3{})
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&logger.ZapLogger{})
	if err != nil {
		t.Fatal(err)
	}

	err = container.Init()
	assert.NoError(t, err)

	_, err = container.Serve()
	assert.Error(t, err)
}

func TestAppWrongCommand(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr-wrong-command.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&server.Plugin{})
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&Foo3{})
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

	_, err = container.Serve()
	assert.Error(t, err)
}

func TestAppNoAppSectionInConfig(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr-wrong-command.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&server.Plugin{})
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&Foo3{})
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

	_, err = container.Serve()
	assert.Error(t, err)
}
