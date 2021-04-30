package informer

import (
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	endure "github.com/spiral/endure/pkg/container"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	"github.com/spiral/roadrunner/v2/pkg/process"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/informer"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/stretchr/testify/assert"
)

func TestInformerInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
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

	stopCh := make(chan struct{}, 1)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				return
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second)
	t.Run("InformerWorkersRpcTest", informerWorkersRPCTest)
	t.Run("InformerListRpcTest", informerListRPCTest)

	stopCh <- struct{}{}
	wg.Wait()
}

func informerWorkersRPCTest(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.
	list := struct {
		// Workers is list of workers.
		Workers []process.State `json:"workers"`
	}{}

	err = client.Call("informer.Workers", "informer.plugin1", &list)
	assert.NoError(t, err)
	assert.Len(t, list.Workers, 10)
}

func informerListRPCTest(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.
	list := make([]string, 0, 5)
	// Plugins which are expected to be in the list
	expected := []string{"rpc", "logs", "informer.plugin1", "config", "server"}

	err = client.Call("informer.List", true, &list)
	assert.NoError(t, err)
	assert.ElementsMatch(t, list, expected)
}
