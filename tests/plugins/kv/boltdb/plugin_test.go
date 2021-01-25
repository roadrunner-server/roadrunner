package boltdb_tests //nolint:golint,stylecheck

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
	"github.com/spiral/roadrunner/v2/plugins/config"
	httpPlugin "github.com/spiral/roadrunner/v2/plugins/http"
	"github.com/spiral/roadrunner/v2/plugins/kv"
	"github.com/spiral/roadrunner/v2/plugins/kv/boltdb"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/stretchr/testify/assert"
)

func TestBoltDb(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-init.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&boltdb.Plugin{},
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

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
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

	time.Sleep(time.Second * 1)
	t.Run("testBoltDbRPCMethods", testRPCMethods)
	stopCh <- struct{}{}
	wg.Wait()

	_ = os.Remove("rr")
}

func testRPCMethods(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	var setRes bool
	items := make([]kv.Item, 0, 5)
	items = append(items, kv.Item{
		Key:   "a",
		Value: "aa",
	})
	items = append(items, kv.Item{
		Key:   "b",
		Value: "bb",
	})
	// add 5 second ttl
	tt := time.Now().Add(time.Second * 5).Format(time.RFC3339)
	items = append(items, kv.Item{
		Key:   "c",
		Value: "cc",
		TTL:   tt,
	})

	items = append(items, kv.Item{
		Key:   "d",
		Value: "dd",
	})

	items = append(items, kv.Item{
		Key:   "e",
		Value: "ee",
	})

	// Register 3 keys with values
	err = client.Call("boltdb.Set", items, &setRes)
	assert.NoError(t, err)
	assert.True(t, setRes)

	ret := make(map[string]bool)
	keys := []string{"a", "b", "c"}
	err = client.Call("boltdb.Has", keys, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 3) // should be 3

	// key "c" should be deleted
	time.Sleep(time.Second * 7)

	ret = make(map[string]bool)
	err = client.Call("boltdb.Has", keys, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 2) // should be 2

	mGet := make(map[string]interface{})
	keys = []string{"a", "b", "c"}
	err = client.Call("boltdb.MGet", keys, &mGet)
	assert.NoError(t, err)
	assert.Len(t, mGet, 2) // c is expired
	assert.Equal(t, string("aa"), mGet["a"].(string))
	assert.Equal(t, string("bb"), mGet["b"].(string))

	mExpKeys := make([]kv.Item, 0, 2)
	tt2 := time.Now().Add(time.Second * 10).Format(time.RFC3339)
	mExpKeys = append(mExpKeys, kv.Item{Key: "a", TTL: tt2})
	mExpKeys = append(mExpKeys, kv.Item{Key: "b", TTL: tt2})
	mExpKeys = append(mExpKeys, kv.Item{Key: "d", TTL: tt2})

	// MEXPIRE
	var mExpRes bool
	err = client.Call("boltdb.MExpire", mExpKeys, &mExpRes)
	assert.NoError(t, err)
	assert.True(t, mExpRes)

	// TTL
	keys = []string{"a", "b", "d"}
	ttlRes := make(map[string]interface{})
	err = client.Call("boltdb.TTL", keys, &ttlRes)
	assert.NoError(t, err)
	assert.Len(t, ttlRes, 3)

	// HAS AFTER TTL
	time.Sleep(time.Second * 15)
	ret = make(map[string]bool)
	keys = []string{"a", "b", "d"}
	err = client.Call("boltdb.Has", keys, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 0)

	// DELETE
	keys = []string{"e"}
	var delRet bool
	err = client.Call("boltdb.Delete", keys, &delRet)
	assert.NoError(t, err)
	assert.True(t, delRet)

	// HAS AFTER DELETE
	ret = make(map[string]bool)
	keys = []string{"e"}
	err = client.Call("boltdb.Has", keys, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 0)
}
