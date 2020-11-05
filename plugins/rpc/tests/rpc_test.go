package tests

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/spiral/endure"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/stretchr/testify/assert"
)

// graph https://bit.ly/3ensdNb
func TestRpcInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel))
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&Plugin1{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&Plugin2{})
	if err != nil {
		t.Fatal(err)
	}

	v := &config.Viper{}
	v.Path = ".rr.yaml"
	v.Prefix = "rr"
	err = cont.Register(v)
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
	tt := time.NewTimer(time.Second * 10)

	for {
		select {
		case e := <-ch:
			// just stop, this is ok
			if errors.Is(errors.Disabled, e.Error) {
				return
			}
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
			assert.Fail(t, "timeout")
		}
	}
}

// graph https://bit.ly/3ensdNb
func TestRpcDisabled(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel))
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&Plugin1{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&Plugin2{})
	if err != nil {
		t.Fatal(err)
	}

	v := &config.Viper{}
	v.Path = ".rr-rpc-disabled.yaml"
	v.Prefix = "rr"
	err = cont.Register(v)
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
	tt := time.NewTimer(time.Second * 20)

	for {
		select {
		case e := <-ch:
			// RPC is turned off, should be and dial error
			if errors.Is(errors.Disabled, e.Error) {
				assert.FailNow(t, "should not be disabled error")
			}
			assert.Error(t, e.Error)
			err = cont.Stop()
			assert.Error(t, err)
			return
		case <-sig:
			err = cont.Stop()
			if err != nil {
				assert.FailNow(t, "error", err.Error())
			}
			return
		case <-tt.C:
			// timeout
			return
		}
	}
}
