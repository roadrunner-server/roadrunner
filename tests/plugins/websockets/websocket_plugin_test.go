package websockets

import (
	"net"
	"net/http"
	"net/rpc"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	json "github.com/json-iterator/go"
	endure "github.com/spiral/endure/pkg/container"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	"github.com/spiral/roadrunner/v2/plugins/channel"
	"github.com/spiral/roadrunner/v2/plugins/config"
	httpPlugin "github.com/spiral/roadrunner/v2/plugins/http"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/memory"
	"github.com/spiral/roadrunner/v2/plugins/redis"
	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/spiral/roadrunner/v2/plugins/websockets"
	"github.com/spiral/roadrunner/v2/utils"
	"github.com/stretchr/testify/assert"
)

type Msg struct {
	// Topic message been pushed into.
	Topics []string `json:"topic"`

	// Command (join, leave, headers)
	Command string `json:"command"`

	// Broker (redis, memory)
	Broker string `json:"broker"`

	// Payload to be broadcasted
	Payload []byte `json:"payload"`
}

func TestBroadcastInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-init.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&redis.Plugin{},
		&websockets.Plugin{},
		&httpPlugin.Plugin{},
		&channel.Plugin{},
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
	t.Run("TestWSInit", wsInit)

	stopCh <- struct{}{}

	wg.Wait()
}

func wsInit(t *testing.T) {
	da := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: time.Second * 20,
	}

	connURL := url.URL{Scheme: "ws", Host: "localhost:11111", Path: "/ws"}

	c, resp, err := da.Dial(connURL.String(), nil)
	assert.NoError(t, err)

	defer func() {
		_ = resp.Body.Close()
	}()

	d, err := json.Marshal(message("join", "memory", []byte("hello websockets"), "foo", "foo2"))
	if err != nil {
		panic(err)
	}

	err = c.WriteMessage(websocket.BinaryMessage, d)
	assert.NoError(t, err)

	_, msg, err := c.ReadMessage()
	retMsg := utils.AsString(msg)
	assert.NoError(t, err)

	// subscription done
	assert.Equal(t, `{"topic":"@join","payload":["foo","foo2"]}`, retMsg)

	err = c.WriteControl(websocket.CloseMessage, nil, time.Time{})
	assert.NoError(t, err)
}

func TestWSRedisAndMemory(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-redis-memory.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&redis.Plugin{},
		&websockets.Plugin{},
		&httpPlugin.Plugin{},
		&memory.Plugin{},
		&channel.Plugin{},
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
	t.Run("RPCWsMemory", RPCWsMemory)
	t.Run("RPCWsRedis", RPCWsRedis)
	t.Run("RPCWsMemoryPubAsync", RPCWsMemoryPubAsync)

	stopCh <- struct{}{}

	wg.Wait()
}

func RPCWsMemoryPubAsync(t *testing.T) {
	da := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: time.Second * 20,
	}

	connURL := url.URL{Scheme: "ws", Host: "localhost:13235", Path: "/ws"}

	c, resp, err := da.Dial(connURL.String(), nil)
	assert.NoError(t, err)

	defer func() {
		_ = resp.Body.Close()
	}()

	d, err := json.Marshal(message("join", "memory", []byte("hello websockets"), "foo", "foo2"))
	if err != nil {
		panic(err)
	}

	err = c.WriteMessage(websocket.BinaryMessage, d)
	assert.NoError(t, err)

	_, msg, err := c.ReadMessage()
	retMsg := utils.AsString(msg)
	assert.NoError(t, err)

	// subscription done
	assert.Equal(t, `{"topic":"@join","payload":["foo","foo2"]}`, retMsg)

	publishAsync("", "memory", "foo")

	// VERIFY a message
	_, msg, err = c.ReadMessage()
	retMsg = utils.AsString(msg)
	assert.NoError(t, err)
	assert.Equal(t, "hello, PHP", retMsg)

	// //// LEAVE foo, foo2 /////////
	d, err = json.Marshal(message("leave", "memory", []byte("hello websockets"), "foo"))
	if err != nil {
		panic(err)
	}

	err = c.WriteMessage(websocket.BinaryMessage, d)
	assert.NoError(t, err)

	_, msg, err = c.ReadMessage()
	retMsg = utils.AsString(msg)
	assert.NoError(t, err)

	// subscription done
	assert.Equal(t, `{"topic":"@leave","payload":["foo"]}`, retMsg)

	// TRY TO PUBLISH TO UNSUBSCRIBED TOPIC
	publishAsync("", "memory", "foo")

	go func() {
		time.Sleep(time.Second * 5)
		publishAsync2("", "memory", "foo2")
	}()

	// should be only message from the subscribed foo2 topic
	_, msg, err = c.ReadMessage()
	retMsg = utils.AsString(msg)
	assert.NoError(t, err)
	assert.Equal(t, "hello, PHP2", retMsg)

	err = c.WriteControl(websocket.CloseMessage, nil, time.Time{})
	assert.NoError(t, err)
}

func RPCWsMemory(t *testing.T) {
	da := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: time.Second * 20,
	}

	connURL := url.URL{Scheme: "ws", Host: "localhost:13235", Path: "/ws"}

	c, resp, err := da.Dial(connURL.String(), nil)
	assert.NoError(t, err)

	defer func() {
		_ = resp.Body.Close()
	}()

	d, err := json.Marshal(message("join", "memory", []byte("hello websockets"), "foo", "foo2"))
	if err != nil {
		panic(err)
	}

	err = c.WriteMessage(websocket.BinaryMessage, d)
	assert.NoError(t, err)

	_, msg, err := c.ReadMessage()
	retMsg := utils.AsString(msg)
	assert.NoError(t, err)

	// subscription done
	assert.Equal(t, `{"topic":"@join","payload":["foo","foo2"]}`, retMsg)

	publish("", "memory", "foo")

	// VERIFY a message
	_, msg, err = c.ReadMessage()
	retMsg = utils.AsString(msg)
	assert.NoError(t, err)
	assert.Equal(t, "hello, PHP", retMsg)

	// //// LEAVE foo, foo2 /////////
	d, err = json.Marshal(message("leave", "memory", []byte("hello websockets"), "foo"))
	if err != nil {
		panic(err)
	}

	err = c.WriteMessage(websocket.BinaryMessage, d)
	assert.NoError(t, err)

	_, msg, err = c.ReadMessage()
	retMsg = utils.AsString(msg)
	assert.NoError(t, err)

	// subscription done
	assert.Equal(t, `{"topic":"@leave","payload":["foo"]}`, retMsg)

	// TRY TO PUBLISH TO UNSUBSCRIBED TOPIC
	publish("", "memory", "foo")

	go func() {
		time.Sleep(time.Second * 5)
		publish2("", "memory", "foo2")
	}()

	// should be only message from the subscribed foo2 topic
	_, msg, err = c.ReadMessage()
	retMsg = utils.AsString(msg)
	assert.NoError(t, err)
	assert.Equal(t, "hello, PHP2", retMsg)

	err = c.WriteControl(websocket.CloseMessage, nil, time.Time{})
	assert.NoError(t, err)
}

func RPCWsRedis(t *testing.T) {
	da := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: time.Second * 20,
	}

	connURL := url.URL{Scheme: "ws", Host: "localhost:13235", Path: "/ws"}

	c, resp, err := da.Dial(connURL.String(), nil)
	assert.NoError(t, err)

	defer func() {
		_ = resp.Body.Close()
	}()

	d, err := json.Marshal(message("join", "redis", []byte("hello websockets"), "foo", "foo2"))
	if err != nil {
		panic(err)
	}

	err = c.WriteMessage(websocket.BinaryMessage, d)
	assert.NoError(t, err)

	_, msg, err := c.ReadMessage()
	retMsg := utils.AsString(msg)
	assert.NoError(t, err)

	// subscription done
	assert.Equal(t, `{"topic":"@join","payload":["foo","foo2"]}`, retMsg)

	publish("", "redis", "foo")

	// VERIFY a message
	_, msg, err = c.ReadMessage()
	retMsg = utils.AsString(msg)
	assert.NoError(t, err)
	assert.Equal(t, "hello, PHP", retMsg)

	// //// LEAVE foo, foo2 /////////
	d, err = json.Marshal(message("leave", "redis", []byte("hello websockets"), "foo"))
	if err != nil {
		panic(err)
	}

	err = c.WriteMessage(websocket.BinaryMessage, d)
	assert.NoError(t, err)

	_, msg, err = c.ReadMessage()
	retMsg = utils.AsString(msg)
	assert.NoError(t, err)

	// subscription done
	assert.Equal(t, `{"topic":"@leave","payload":["foo"]}`, retMsg)

	// TRY TO PUBLISH TO UNSUBSCRIBED TOPIC
	publish("", "redis", "foo")

	go func() {
		time.Sleep(time.Second * 5)
		publish2("", "redis", "foo2")
	}()

	// should be only message from the subscribed foo2 topic
	_, msg, err = c.ReadMessage()
	retMsg = utils.AsString(msg)
	assert.NoError(t, err)
	assert.Equal(t, "hello, PHP2", retMsg)

	err = c.WriteControl(websocket.CloseMessage, nil, time.Time{})
	assert.NoError(t, err)
}

func publish(command string, broker string, topics ...string) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	var ret bool
	err = client.Call("websockets.Publish", []*Msg{message(command, broker, []byte("hello, PHP"), topics...)}, &ret)
	if err != nil {
		panic(err)
	}
}

func publishAsync(command string, broker string, topics ...string) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	var ret bool
	err = client.Call("websockets.PublishAsync", []*Msg{message(command, broker, []byte("hello, PHP"), topics...)}, &ret)
	if err != nil {
		panic(err)
	}
}

func publishAsync2(command string, broker string, topics ...string) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	var ret bool
	err = client.Call("websockets.PublishAsync", []*Msg{message(command, broker, []byte("hello, PHP2"), topics...)}, &ret)
	if err != nil {
		panic(err)
	}
}

func publish2(command string, broker string, topics ...string) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	var ret bool
	err = client.Call("websockets.Publish", []*Msg{message(command, broker, []byte("hello, PHP2"), topics...)}, &ret)
	if err != nil {
		panic(err)
	}
}

func message(command string, broker string, payload []byte, topics ...string) *Msg {
	return &Msg{
		Topics:  topics,
		Command: command,
		Broker:  broker,
		Payload: payload,
	}
}
