package kv

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
	"github.com/spiral/roadrunner/v2/plugins/kv"
	"github.com/spiral/roadrunner/v2/plugins/kv/drivers/boltdb"
	"github.com/spiral/roadrunner/v2/plugins/kv/drivers/memcached"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/memory"
	"github.com/spiral/roadrunner/v2/plugins/redis"
	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"
	payload "github.com/spiral/roadrunner/v2/proto/kv/v1beta"
	"github.com/stretchr/testify/assert"
)

func TestKVInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-kv-init.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&memory.Plugin{},
		&boltdb.Plugin{},
		&memcached.Plugin{},
		&redis.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&kv.Plugin{},
	)
	assert.NoError(t, err)

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
	t.Run("KvSetTest", kvSetTest)
	t.Run("KvHasTest", kvHasTest)

	stopCh <- struct{}{}

	wg.Wait()

	_ = os.RemoveAll("rr.db")
	_ = os.RemoveAll("africa.db")
}

func TestKVNoInterval(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-kv-bolt-no-interval.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&boltdb.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&kv.Plugin{},
	)
	assert.NoError(t, err)

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
	t.Run("KvSetTest", kvSetTest)
	t.Run("KvHasTest", kvHasTest)

	stopCh <- struct{}{}

	wg.Wait()

	_ = os.RemoveAll("rr.db")
	_ = os.RemoveAll("africa.db")
}

func TestKVCreateToReopenWithPerms(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-kv-bolt-perms.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&boltdb.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&kv.Plugin{},
	)
	assert.NoError(t, err)

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
	stopCh <- struct{}{}
	wg.Wait()
}

func TestKVCreateToReopenWithPerms2(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-kv-bolt-perms.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&boltdb.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&kv.Plugin{},
	)
	assert.NoError(t, err)

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
	t.Run("KvSetTest", kvSetTest)
	t.Run("KvHasTest", kvHasTest)

	stopCh <- struct{}{}

	wg.Wait()

	_ = os.RemoveAll("rr.db")
	_ = os.RemoveAll("africa.db")
}

func kvSetTest(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.
	p := &payload.Request{
		Storage: "boltdb-south",
		Items: []*payload.Item{
			{
				Key:   "key",
				Value: []byte("val"),
			},
		},
	}

	resp := &payload.Response{}
	err = client.Call("kv.Set", p, resp)
	assert.NoError(t, err)
}

func kvHasTest(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.
	p := &payload.Request{
		Storage: "boltdb-south",
		Items: []*payload.Item{
			{
				Key:   "key",
				Value: []byte("val"),
			},
		},
	}

	ret := &payload.Response{}
	err = client.Call("kv.Has", p, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 1)
}

func TestBoltDb(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-boltdb.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&kv.Plugin{},
		&boltdb.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&memory.Plugin{},
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
	t.Run("BOLTDB", testRPCMethods)
	stopCh <- struct{}{}
	wg.Wait()

	_ = os.Remove("rr.db")
}

func testRPCMethods(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	// add 5 second ttl
	tt := time.Now().Add(time.Second * 5).Format(time.RFC3339)
	keys := &payload.Request{
		Storage: "boltdb-rr",
		Items: []*payload.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "c",
			},
		},
	}

	data := &payload.Request{
		Storage: "boltdb-rr",
		Items: []*payload.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:     "c",
				Value:   []byte("cc"),
				Timeout: tt,
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	ret := &payload.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", data, ret)
	assert.NoError(t, err)

	ret = &payload.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 3) // should be 3

	// key "c" should be deleted
	time.Sleep(time.Second * 7)

	ret = &payload.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // should be 2

	ret = &payload.Response{}
	err = client.Call("kv.MGet", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // c is expired

	tt2 := time.Now().Add(time.Second * 10).Format(time.RFC3339)

	data2 := &payload.Request{
		Storage: "boltdb-rr",
		Items: []*payload.Item{
			{
				Key:     "a",
				Timeout: tt2,
			},
			{
				Key:     "b",
				Timeout: tt2,
			},
			{
				Key:     "d",
				Timeout: tt2,
			},
		},
	}

	// MEXPIRE
	ret = &payload.Response{}
	err = client.Call("kv.MExpire", data2, ret)
	assert.NoError(t, err)

	// TTL
	keys2 := &payload.Request{
		Storage: "boltdb-rr",
		Items: []*payload.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "d",
			},
		},
	}

	ret = &payload.Response{}
	err = client.Call("kv.TTL", keys2, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 3)

	// HAS AFTER TTL
	time.Sleep(time.Second * 15)
	ret = &payload.Response{}
	err = client.Call("kv.Has", keys2, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	// DELETE
	keysDel := &payload.Request{
		Storage: "boltdb-rr",
		Items: []*payload.Item{
			{
				Key: "e",
			},
		},
	}

	ret = &payload.Response{}
	err = client.Call("kv.Delete", keysDel, ret)
	assert.NoError(t, err)

	// HAS AFTER DELETE
	ret = &payload.Response{}
	err = client.Call("kv.Has", keysDel, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	dataClear := &payload.Request{
		Storage: "boltdb-rr",
		Items: []*payload.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:   "c",
				Value: []byte("cc"),
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	clear := &payload.Request{Storage: "boltdb-rr"}

	ret = &payload.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", dataClear, ret)
	assert.NoError(t, err)

	ret = &payload.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 5) // should be 5

	ret = &payload.Response{}
	err = client.Call("kv.Clear", clear, ret)
	assert.NoError(t, err)

	ret = &payload.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0) // should be 5
}

func TestMemcached(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-memcached.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&kv.Plugin{},
		&memcached.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&memory.Plugin{},
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
	t.Run("MEMCACHED", testRPCMethodsMemcached)
	stopCh <- struct{}{}
	wg.Wait()
}

func testRPCMethodsMemcached(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	// add 5 second ttl
	tt := time.Now().Add(time.Second * 5).Format(time.RFC3339)

	keys := &payload.Request{
		Storage: "memcached-rr",
		Items: []*payload.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "c",
			},
		},
	}

	data := &payload.Request{
		Storage: "memcached-rr",
		Items: []*payload.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:     "c",
				Value:   []byte("cc"),
				Timeout: tt,
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	ret := &payload.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", data, ret)
	assert.NoError(t, err)

	ret = &payload.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 3) // should be 3

	// key "c" should be deleted
	time.Sleep(time.Second * 7)

	ret = &payload.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // should be 2

	ret = &payload.Response{}
	err = client.Call("kv.MGet", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // c is expired

	tt2 := time.Now().Add(time.Second * 10).Format(time.RFC3339)

	data2 := &payload.Request{
		Storage: "memcached-rr",
		Items: []*payload.Item{
			{
				Key:     "a",
				Timeout: tt2,
			},
			{
				Key:     "b",
				Timeout: tt2,
			},
			{
				Key:     "d",
				Timeout: tt2,
			},
		},
	}

	// MEXPIRE
	ret = &payload.Response{}
	err = client.Call("kv.MExpire", data2, ret)
	assert.NoError(t, err)

	// TTL call is not supported for the memcached driver
	keys2 := &payload.Request{
		Storage: "memcached-rr",
		Items: []*payload.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "d",
			},
		},
	}

	ret = &payload.Response{}
	err = client.Call("kv.TTL", keys2, ret)
	assert.Error(t, err)
	assert.Len(t, ret.GetItems(), 0)

	// HAS AFTER TTL
	time.Sleep(time.Second * 15)
	ret = &payload.Response{}
	err = client.Call("kv.Has", keys2, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	// DELETE
	keysDel := &payload.Request{
		Storage: "memcached-rr",
		Items: []*payload.Item{
			{
				Key: "e",
			},
		},
	}

	ret = &payload.Response{}
	err = client.Call("kv.Delete", keysDel, ret)
	assert.NoError(t, err)

	// HAS AFTER DELETE
	ret = &payload.Response{}
	err = client.Call("kv.Has", keysDel, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	dataClear := &payload.Request{
		Storage: "memcached-rr",
		Items: []*payload.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:   "c",
				Value: []byte("cc"),
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	clear := &payload.Request{Storage: "memcached-rr"}

	ret = &payload.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", dataClear, ret)
	assert.NoError(t, err)

	ret = &payload.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 5) // should be 5

	ret = &payload.Response{}
	err = client.Call("kv.Clear", clear, ret)
	assert.NoError(t, err)

	time.Sleep(time.Second * 2)
	ret = &payload.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0) // should be 5
}

func TestInMemory(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-in-memory.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&kv.Plugin{},
		&memory.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
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
	t.Run("INMEMORY", testRPCMethodsInMemory)
	stopCh <- struct{}{}
	wg.Wait()
}

func testRPCMethodsInMemory(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	// add 5 second ttl

	tt := time.Now().Add(time.Second * 5).Format(time.RFC3339)
	keys := &payload.Request{
		Storage: "memory-rr",
		Items: []*payload.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "c",
			},
		},
	}

	data := &payload.Request{
		Storage: "memory-rr",
		Items: []*payload.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:     "c",
				Value:   []byte("cc"),
				Timeout: tt,
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	ret := &payload.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", data, ret)
	assert.NoError(t, err)

	ret = &payload.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 3) // should be 3

	// key "c" should be deleted
	time.Sleep(time.Second * 7)

	ret = &payload.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // should be 2

	ret = &payload.Response{}
	err = client.Call("kv.MGet", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // c is expired

	tt2 := time.Now().Add(time.Second * 10).Format(time.RFC3339)

	data2 := &payload.Request{
		Storage: "memory-rr",
		Items: []*payload.Item{
			{
				Key:     "a",
				Timeout: tt2,
			},
			{
				Key:     "b",
				Timeout: tt2,
			},
			{
				Key:     "d",
				Timeout: tt2,
			},
		},
	}

	// MEXPIRE
	ret = &payload.Response{}
	err = client.Call("kv.MExpire", data2, ret)
	assert.NoError(t, err)

	// TTL
	keys2 := &payload.Request{
		Storage: "memory-rr",
		Items: []*payload.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "d",
			},
		},
	}

	ret = &payload.Response{}
	err = client.Call("kv.TTL", keys2, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 3)

	// HAS AFTER TTL
	time.Sleep(time.Second * 15)
	ret = &payload.Response{}
	err = client.Call("kv.Has", keys2, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	// DELETE
	keysDel := &payload.Request{
		Storage: "memory-rr",
		Items: []*payload.Item{
			{
				Key: "e",
			},
		},
	}

	ret = &payload.Response{}
	err = client.Call("kv.Delete", keysDel, ret)
	assert.NoError(t, err)

	// HAS AFTER DELETE
	ret = &payload.Response{}
	err = client.Call("kv.Has", keysDel, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	dataClear := &payload.Request{
		Storage: "memory-rr",
		Items: []*payload.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:   "c",
				Value: []byte("cc"),
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	clear := &payload.Request{Storage: "memory-rr"}

	ret = &payload.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", dataClear, ret)
	assert.NoError(t, err)

	ret = &payload.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 5) // should be 5

	ret = &payload.Response{}
	err = client.Call("kv.Clear", clear, ret)
	assert.NoError(t, err)

	ret = &payload.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0) // should be 5
}

func TestRedis(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-redis.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&kv.Plugin{},
		&redis.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&memory.Plugin{},
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
	t.Run("REDIS", testRPCMethodsRedis)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestRedisGlobalSection(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-redis-global.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&kv.Plugin{},
		&redis.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&memory.Plugin{},
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
	t.Run("REDIS", testRPCMethodsRedis)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestRedisNoConfig(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-redis-no-config.yaml", // should be used default
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&kv.Plugin{},
		&redis.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&memory.Plugin{},
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
	t.Run("REDIS", testRPCMethodsRedis)
	stopCh <- struct{}{}
	wg.Wait()
}

func testRPCMethodsRedis(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	// add 5 second ttl
	tt := time.Now().Add(time.Second * 5).Format(time.RFC3339)
	keys := &payload.Request{
		Storage: "redis-rr",
		Items: []*payload.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "c",
			},
		},
	}

	data := &payload.Request{
		Storage: "redis-rr",
		Items: []*payload.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:     "c",
				Value:   []byte("cc"),
				Timeout: tt,
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	ret := &payload.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", data, ret)
	assert.NoError(t, err)

	ret = &payload.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 3) // should be 3

	// key "c" should be deleted
	time.Sleep(time.Second * 7)

	ret = &payload.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // should be 2

	ret = &payload.Response{}
	err = client.Call("kv.MGet", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // c is expired

	tt2 := time.Now().Add(time.Second * 10).Format(time.RFC3339)

	data2 := &payload.Request{
		Storage: "redis-rr",
		Items: []*payload.Item{
			{
				Key:     "a",
				Timeout: tt2,
			},
			{
				Key:     "b",
				Timeout: tt2,
			},
			{
				Key:     "d",
				Timeout: tt2,
			},
		},
	}

	// MEXPIRE
	ret = &payload.Response{}
	err = client.Call("kv.MExpire", data2, ret)
	assert.NoError(t, err)

	// TTL
	keys2 := &payload.Request{
		Storage: "redis-rr",
		Items: []*payload.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "d",
			},
		},
	}

	ret = &payload.Response{}
	err = client.Call("kv.TTL", keys2, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 3)

	// HAS AFTER TTL
	time.Sleep(time.Second * 15)
	ret = &payload.Response{}
	err = client.Call("kv.Has", keys2, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	// DELETE
	keysDel := &payload.Request{
		Storage: "redis-rr",
		Items: []*payload.Item{
			{
				Key: "e",
			},
		},
	}

	ret = &payload.Response{}
	err = client.Call("kv.Delete", keysDel, ret)
	assert.NoError(t, err)

	// HAS AFTER DELETE
	ret = &payload.Response{}
	err = client.Call("kv.Has", keysDel, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	dataClear := &payload.Request{
		Storage: "redis-rr",
		Items: []*payload.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:   "c",
				Value: []byte("cc"),
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	clear := &payload.Request{Storage: "redis-rr"}

	ret = &payload.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", dataClear, ret)
	assert.NoError(t, err)

	ret = &payload.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 5) // should be 5

	ret = &payload.Response{}
	err = client.Call("kv.Clear", clear, ret)
	assert.NoError(t, err)

	ret = &payload.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0) // should be 5
}
