package tests

import (
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/spiral/endure"
	"github.com/spiral/goridge/v3"
	"github.com/spiral/roadrunner/v2"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/informer"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/stretchr/testify/assert"
)

func TestInformerInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Viper{
		Path:   ".rr-informer.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&logger.ZapLogger{},
		&informer.Plugin{},
		&rpcPlugin.Plugin{},
		&Plugin1{},
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

	tt := time.NewTimer(time.Second * 15)

	t.Run("InformerRpcTest", informerRPCTest)

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

func informerRPCTest(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridge.NewClientCodec(conn))
	// WorkerList contains list of workers.
	list := struct {
		// Workers is list of workers.
		Workers []roadrunner.ProcessState `json:"workers"`
	}{}

	err = client.Call("informer.Workers", "informer.plugin1", &list)
	assert.NoError(t, err)
	assert.Len(t, list.Workers, 10)
}
