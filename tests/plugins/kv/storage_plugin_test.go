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
	"github.com/spiral/roadrunner/v2/plugins/kv/payload/generated"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/redis"
	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/stretchr/testify/assert"
)

func makePayload(b *flatbuffers.Builder, storage string, items []kv.Item) []byte {
	b.Reset()

	storageOffset := b.CreateString(storage)

	////////////////////// ITEMS VECTOR ////////////////////////////
	offset := make([]flatbuffers.UOffsetT, len(items))
	for i := len(items) - 1; i >= 0; i-- {
		offset[i] = serializeItems(b, items[i])
	}

	generated.PayloadStartItemsVector(b, len(offset))

	for i := len(offset) - 1; i >= 0; i-- {
		b.PrependUOffsetT(offset[i])
	}

	itemsOffset := b.EndVector(len(offset))
	///////////////////////////////////////////////////////////////////

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
