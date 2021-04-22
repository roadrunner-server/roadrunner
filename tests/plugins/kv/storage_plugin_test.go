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

	flatbuffers "github.com/google/flatbuffers/go"
	endure "github.com/spiral/endure/pkg/container"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/kv"
	"github.com/spiral/roadrunner/v2/plugins/kv/drivers/boltdb"
	"github.com/spiral/roadrunner/v2/plugins/kv/drivers/memcached"
	"github.com/spiral/roadrunner/v2/plugins/kv/drivers/memory"
	"github.com/spiral/roadrunner/v2/plugins/kv/drivers/redis"
	"github.com/spiral/roadrunner/v2/plugins/kv/payload/generated"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/stretchr/testify/assert"
)

func makePayload(b *flatbuffers.Builder, storage string, items []kv.Item) []byte {
	b.Reset()

	storageOffset := b.CreateString(storage)

	// //////////////////// ITEMS VECTOR ////////////////////////////
	offset := make([]flatbuffers.UOffsetT, len(items))
	for i := len(items) - 1; i >= 0; i-- {
		offset[i] = serializeItems(b, items[i])
	}

	generated.PayloadStartItemsVector(b, len(offset))

	for i := len(offset) - 1; i >= 0; i-- {
		b.PrependUOffsetT(offset[i])
	}

	itemsOffset := b.EndVector(len(offset))
	// /////////////////////////////////////////////////////////////////

	generated.PayloadStart(b)
	generated.PayloadAddItems(b, itemsOffset)
	generated.PayloadAddStorage(b, storageOffset)

	finalOffset := generated.PayloadEnd(b)

	b.Finish(finalOffset)

	return b.Bytes[b.Head():]
}

func serializeItems(b *flatbuffers.Builder, item kv.Item) flatbuffers.UOffsetT {
	key := b.CreateString(item.Key)
	val := b.CreateString(item.Value)
	ttl := b.CreateString(item.TTL)

	generated.ItemStart(b)

	generated.ItemAddKey(b, key)
	generated.ItemAddValue(b, val)
	generated.ItemAddTimeout(b, ttl)

	return generated.ItemEnd(b)
}

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

func kvSetTest(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.

	b := flatbuffers.NewBuilder(100)
	args := makePayload(b, "boltdb-south", []kv.Item{
		{
			Key:   "key",
			Value: "val",
		},
	})

	var ok bool

	err = client.Call("kv.Set", args, &ok)
	assert.NoError(t, err)
	assert.True(t, ok, "Set return result")
}

func kvHasTest(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.

	b := flatbuffers.NewBuilder(100)
	args := makePayload(b, "boltdb-south", []kv.Item{
		{
			Key:   "key",
			Value: "val",
		},
	})
	var ret map[string]bool

	err = client.Call("kv.Has", args, &ret)
	assert.NoError(t, err)
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

	_ = os.Remove("rr.db")
}

func testRPCMethods(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	// add 5 second ttl
	tt := time.Now().Add(time.Second * 5).Format(time.RFC3339)
	keys := makePayload(flatbuffers.NewBuilder(100), "boltdb-rr", []kv.Item{
		{
			Key: "a",
		},
		{
			Key: "b",
		},
		{
			Key: "c",
		},
	})
	data := makePayload(flatbuffers.NewBuilder(100), "boltdb-rr", []kv.Item{
		{
			Key:   "a",
			Value: "aa",
		},
		{
			Key:   "b",
			Value: "bb",
		},
		{
			Key:   "c",
			Value: "cc",
			TTL:   tt,
		},
		{
			Key:   "d",
			Value: "dd",
		},
		{
			Key:   "e",
			Value: "ee",
		},
	})

	var setRes bool

	// Register 3 keys with values
	err = client.Call("kv.Set", data, &setRes)
	assert.NoError(t, err)
	assert.True(t, setRes)

	ret := make(map[string]bool)
	err = client.Call("kv.Has", keys, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 3) // should be 3

	// key "c" should be deleted
	time.Sleep(time.Second * 7)

	ret = make(map[string]bool)
	err = client.Call("kv.Has", keys, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 2) // should be 2

	mGet := make(map[string]interface{})
	err = client.Call("kv.MGet", keys, &mGet)
	assert.NoError(t, err)
	assert.Len(t, mGet, 2) // c is expired
	assert.Equal(t, "aa", mGet["a"].(string))
	assert.Equal(t, "bb", mGet["b"].(string))

	tt2 := time.Now().Add(time.Second * 10).Format(time.RFC3339)
	data2 := makePayload(flatbuffers.NewBuilder(100), "boltdb-rr", []kv.Item{
		{
			Key: "a",
			TTL: tt2,
		},
		{
			Key: "b",
			TTL: tt2,
		},
		{
			Key: "d",
			TTL: tt2,
		},
	})

	// MEXPIRE
	var mExpRes bool
	err = client.Call("kv.MExpire", data2, &mExpRes)
	assert.NoError(t, err)
	assert.True(t, mExpRes)

	// TTL
	keys2 := makePayload(flatbuffers.NewBuilder(100), "boltdb-rr", []kv.Item{
		{
			Key: "a",
		},
		{
			Key: "b",
		},
		{
			Key: "d",
		},
	})
	ttlRes := make(map[string]interface{})
	err = client.Call("kv.TTL", keys2, &ttlRes)
	assert.NoError(t, err)
	assert.Len(t, ttlRes, 3)

	// HAS AFTER TTL
	time.Sleep(time.Second * 15)
	ret = make(map[string]bool)
	err = client.Call("kv.Has", keys2, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 0)

	// DELETE
	keysDel := makePayload(flatbuffers.NewBuilder(100), "boltdb-rr", []kv.Item{
		{
			Key: "e",
		},
	})
	var delRet bool
	err = client.Call("kv.Delete", keysDel, &delRet)
	assert.NoError(t, err)
	assert.True(t, delRet)

	// HAS AFTER DELETE
	ret = make(map[string]bool)
	err = client.Call("kv.Has", keysDel, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 0)
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
	t.Run("testMemcachedRPCMethods", testRPCMethodsMemcached)
	stopCh <- struct{}{}
	wg.Wait()
}

func testRPCMethodsMemcached(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	// add 5 second ttl
	tt := time.Now().Add(time.Second * 5).Format(time.RFC3339)
	keys := makePayload(flatbuffers.NewBuilder(100), "memcached-rr", []kv.Item{
		{
			Key: "a",
		},
		{
			Key: "b",
		},
		{
			Key: "c",
		},
	})
	data := makePayload(flatbuffers.NewBuilder(100), "memcached-rr", []kv.Item{
		{
			Key:   "a",
			Value: "aa",
		},
		{
			Key:   "b",
			Value: "bb",
		},
		{
			Key:   "c",
			Value: "cc",
			TTL:   tt,
		},
		{
			Key:   "d",
			Value: "dd",
		},
		{
			Key:   "e",
			Value: "ee",
		},
	})

	var setRes bool

	// Register 3 keys with values
	err = client.Call("kv.Set", data, &setRes)
	assert.NoError(t, err)
	assert.True(t, setRes)

	ret := make(map[string]bool)
	err = client.Call("kv.Has", keys, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 3) // should be 3

	// key "c" should be deleted
	time.Sleep(time.Second * 7)

	ret = make(map[string]bool)
	err = client.Call("kv.Has", keys, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 2) // should be 2

	mGet := make(map[string]interface{})
	err = client.Call("kv.MGet", keys, &mGet)
	assert.NoError(t, err)
	assert.Len(t, mGet, 2) // c is expired
	assert.Equal(t, string("aa"), string(mGet["a"].([]byte)))
	assert.Equal(t, string("bb"), string(mGet["b"].([]byte)))

	tt2 := time.Now().Add(time.Second * 10).Format(time.RFC3339)
	data2 := makePayload(flatbuffers.NewBuilder(100), "memcached-rr", []kv.Item{
		{
			Key: "a",
			TTL: tt2,
		},
		{
			Key: "b",
			TTL: tt2,
		},
		{
			Key: "d",
			TTL: tt2,
		},
	})

	// MEXPIRE
	var mExpRes bool
	err = client.Call("kv.MExpire", data2, &mExpRes)
	assert.NoError(t, err)
	assert.True(t, mExpRes)

	// TTL call is not supported for the memcached driver
	keys2 := makePayload(flatbuffers.NewBuilder(100), "memcached-rr", []kv.Item{
		{
			Key: "a",
		},
		{
			Key: "b",
		},
		{
			Key: "d",
		},
	})
	ttlRes := make(map[string]interface{})
	err = client.Call("kv.TTL", keys2, &ttlRes)
	assert.Error(t, err)
	assert.Len(t, ttlRes, 0)

	// HAS AFTER TTL
	time.Sleep(time.Second * 15)
	ret = make(map[string]bool)
	err = client.Call("kv.Has", keys2, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 0)

	// DELETE
	keysDel := makePayload(flatbuffers.NewBuilder(100), "memcached-rr", []kv.Item{
		{
			Key: "e",
		},
	})
	var delRet bool
	err = client.Call("kv.Delete", keysDel, &delRet)
	assert.NoError(t, err)
	assert.True(t, delRet)

	// HAS AFTER DELETE
	ret = make(map[string]bool)
	err = client.Call("kv.Has", keysDel, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 0)
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
	t.Run("testInMemoryRPCMethods", testRPCMethodsInMemory)
	stopCh <- struct{}{}
	wg.Wait()
}

func testRPCMethodsInMemory(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	// add 5 second ttl
	tt := time.Now().Add(time.Second * 5).Format(time.RFC3339)
	keys := makePayload(flatbuffers.NewBuilder(100), "memory-rr", []kv.Item{
		{
			Key: "a",
		},
		{
			Key: "b",
		},
		{
			Key: "c",
		},
	})
	data := makePayload(flatbuffers.NewBuilder(100), "memory-rr", []kv.Item{
		{
			Key:   "a",
			Value: "aa",
		},
		{
			Key:   "b",
			Value: "bb",
		},
		{
			Key:   "c",
			Value: "cc",
			TTL:   tt,
		},
		{
			Key:   "d",
			Value: "dd",
		},
		{
			Key:   "e",
			Value: "ee",
		},
	})

	var setRes bool

	// Register 3 keys with values
	err = client.Call("kv.Set", data, &setRes)
	assert.NoError(t, err)
	assert.True(t, setRes)

	ret := make(map[string]bool)
	err = client.Call("kv.Has", keys, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 3) // should be 3

	// key "c" should be deleted
	time.Sleep(time.Second * 7)

	ret = make(map[string]bool)
	err = client.Call("kv.Has", keys, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 2) // should be 2

	mGet := make(map[string]interface{})
	err = client.Call("kv.MGet", keys, &mGet)
	assert.NoError(t, err)
	assert.Len(t, mGet, 2) // c is expired
	assert.Equal(t, "aa", mGet["a"].(string))
	assert.Equal(t, "bb", mGet["b"].(string))

	tt2 := time.Now().Add(time.Second * 10).Format(time.RFC3339)
	data2 := makePayload(flatbuffers.NewBuilder(100), "memory-rr", []kv.Item{
		{
			Key: "a",
			TTL: tt2,
		},
		{
			Key: "b",
			TTL: tt2,
		},
		{
			Key: "d",
			TTL: tt2,
		},
	})

	// MEXPIRE
	var mExpRes bool
	err = client.Call("kv.MExpire", data2, &mExpRes)
	assert.NoError(t, err)
	assert.True(t, mExpRes)

	// TTL
	keys2 := makePayload(flatbuffers.NewBuilder(100), "memory-rr", []kv.Item{
		{
			Key: "a",
		},
		{
			Key: "b",
		},
		{
			Key: "d",
		},
	})
	ttlRes := make(map[string]interface{})
	err = client.Call("kv.TTL", keys2, &ttlRes)
	assert.NoError(t, err)
	assert.Len(t, ttlRes, 3)

	// HAS AFTER TTL
	time.Sleep(time.Second * 15)
	ret = make(map[string]bool)
	err = client.Call("kv.Has", keys2, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 0)

	// DELETE
	keysDel := makePayload(flatbuffers.NewBuilder(100), "memory-rr", []kv.Item{
		{
			Key: "e",
		},
	})
	var delRet bool
	err = client.Call("kv.Delete", keysDel, &delRet)
	assert.NoError(t, err)
	assert.True(t, delRet)

	// HAS AFTER DELETE
	ret = make(map[string]bool)
	err = client.Call("kv.Has", keysDel, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 0)
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
	t.Run("testRedisRPCMethods", testRPCMethodsRedis)
	stopCh <- struct{}{}
	wg.Wait()
}

func testRPCMethodsRedis(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	// add 5 second ttl
	tt := time.Now().Add(time.Second * 5).Format(time.RFC3339)
	keys := makePayload(flatbuffers.NewBuilder(100), "redis-rr", []kv.Item{
		{
			Key: "a",
		},
		{
			Key: "b",
		},
		{
			Key: "c",
		},
	})
	data := makePayload(flatbuffers.NewBuilder(100), "redis-rr", []kv.Item{
		{
			Key:   "a",
			Value: "aa",
		},
		{
			Key:   "b",
			Value: "bb",
		},
		{
			Key:   "c",
			Value: "cc",
			TTL:   tt,
		},
		{
			Key:   "d",
			Value: "dd",
		},
		{
			Key:   "e",
			Value: "ee",
		},
	})

	var setRes bool

	// Register 3 keys with values
	err = client.Call("kv.Set", data, &setRes)
	assert.NoError(t, err)
	assert.True(t, setRes)

	ret := make(map[string]bool)
	err = client.Call("kv.Has", keys, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 3) // should be 3

	// key "c" should be deleted
	time.Sleep(time.Second * 7)

	ret = make(map[string]bool)
	err = client.Call("kv.Has", keys, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 2) // should be 2

	mGet := make(map[string]interface{})
	err = client.Call("kv.MGet", keys, &mGet)
	assert.NoError(t, err)
	assert.Len(t, mGet, 2) // c is expired
	assert.Equal(t, "aa", mGet["a"].(string))
	assert.Equal(t, "bb", mGet["b"].(string))

	tt2 := time.Now().Add(time.Second * 10).Format(time.RFC3339)
	data2 := makePayload(flatbuffers.NewBuilder(100), "redis-rr", []kv.Item{
		{
			Key: "a",
			TTL: tt2,
		},
		{
			Key: "b",
			TTL: tt2,
		},
		{
			Key: "d",
			TTL: tt2,
		},
	})

	// MEXPIRE
	var mExpRes bool
	err = client.Call("kv.MExpire", data2, &mExpRes)
	assert.NoError(t, err)
	assert.True(t, mExpRes)

	// TTL
	keys2 := makePayload(flatbuffers.NewBuilder(100), "redis-rr", []kv.Item{
		{
			Key: "a",
		},
		{
			Key: "b",
		},
		{
			Key: "d",
		},
	})
	ttlRes := make(map[string]interface{})
	err = client.Call("kv.TTL", keys2, &ttlRes)
	assert.NoError(t, err)
	assert.Len(t, ttlRes, 3)

	// HAS AFTER TTL
	time.Sleep(time.Second * 15)
	ret = make(map[string]bool)
	err = client.Call("kv.Has", keys2, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 0)

	// DELETE
	keysDel := makePayload(flatbuffers.NewBuilder(100), "redis-rr", []kv.Item{
		{
			Key: "e",
		},
	})
	var delRet bool
	err = client.Call("kv.Delete", keysDel, &delRet)
	assert.NoError(t, err)
	assert.True(t, delRet)

	// HAS AFTER DELETE
	ret = make(map[string]bool)
	err = client.Call("kv.Has", keysDel, &ret)
	assert.NoError(t, err)
	assert.Len(t, ret, 0)
}
