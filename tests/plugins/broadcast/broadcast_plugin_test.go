package broadcast

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
	websocketsv1 "github.com/spiral/roadrunner/v2/pkg/proto/websockets/v1beta"
	"github.com/spiral/roadrunner/v2/plugins/broadcast"
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

func TestBroadcastInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-broadcast-init.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&broadcast.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&redis.Plugin{},
		&websockets.Plugin{},
		&httpPlugin.Plugin{},
		&memory.Plugin{},
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
	//t.Run("TestWSInit", wsInit)

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

	d, err := json.Marshal(messageWS("join", "memory", []byte("hello websockets"), "foo", "foo2"))
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

func publishAsync(t *testing.T, command string, broker string, topics ...string) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	ret := &websocketsv1.Response{}
	err = client.Call("websockets.PublishAsync", makeMessage(command, broker, []byte("hello, PHP"), topics...), ret)
	assert.NoError(t, err)
	assert.True(t, ret.Ok)
}

func publishAsync2(t *testing.T, command string, broker string, topics ...string) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	ret := &websocketsv1.Response{}
	err = client.Call("websockets.PublishAsync", makeMessage(command, broker, []byte("hello, PHP2"), topics...), ret)
	assert.NoError(t, err)
	assert.True(t, ret.Ok)
}

func publish2(t *testing.T, command string, broker string, topics ...string) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	ret := &websocketsv1.Response{}
	err = client.Call("websockets.Publish", makeMessage(command, broker, []byte("hello, PHP2"), topics...), ret)
	assert.NoError(t, err)
	assert.True(t, ret.Ok)
}

func messageWS(command string, broker string, payload []byte, topics ...string) *websocketsv1.Message {
	return &websocketsv1.Message{
		Topics:  topics,
		Command: command,
		Broker:  broker,
		Payload: payload,
	}
}

func makeMessage(command string, broker string, payload []byte, topics ...string) *websocketsv1.Request {
	m := &websocketsv1.Request{
		Messages: []*websocketsv1.Message{
			{
				Topics:  topics,
				Command: command,
				Broker:  broker,
				Payload: payload,
			},
		},
	}

	return m
}
