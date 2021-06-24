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
	"github.com/spiral/roadrunner/v2/plugins/broadcast"
	"github.com/spiral/roadrunner/v2/plugins/config"
	httpPlugin "github.com/spiral/roadrunner/v2/plugins/http"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/memory"
	"github.com/spiral/roadrunner/v2/plugins/redis"
	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/spiral/roadrunner/v2/plugins/websockets"
	websocketsv1 "github.com/spiral/roadrunner/v2/proto/websockets/v1beta"
	"github.com/spiral/roadrunner/v2/utils"
	"github.com/stretchr/testify/assert"
)

func TestWebsocketsInit(t *testing.T) {
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
		&memory.Plugin{},
		&broadcast.Plugin{},
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
	t.Run("RPCWsMemoryPubAsync", RPCWsPubAsync("11111"))
	t.Run("RPCWsMemory", RPCWsPub("11111"))

	stopCh <- struct{}{}

	wg.Wait()
}

func TestWSRedis(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-redis.yaml",
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
		&broadcast.Plugin{},
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
	t.Run("RPCWsRedisPubAsync", RPCWsPubAsync("13235"))
	t.Run("RPCWsRedisPub", RPCWsPub("13235"))

	stopCh <- struct{}{}

	wg.Wait()
}

func TestWSRedisNoSection(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-broker-no-section.yaml",
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
		&broadcast.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	_, err = cont.Serve()
	assert.Error(t, err)
}

func TestWSDeny(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-deny.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&websockets.Plugin{},
		&httpPlugin.Plugin{},
		&memory.Plugin{},
		&broadcast.Plugin{},
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
	t.Run("RPCWsMemoryDeny", RPCWsDeny("15587"))

	stopCh <- struct{}{}

	wg.Wait()
}

func TestWSDeny2(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-deny2.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&websockets.Plugin{},
		&httpPlugin.Plugin{},
		&redis.Plugin{},
		&broadcast.Plugin{},
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
	t.Run("RPCWsRedisDeny", RPCWsDeny("15588"))

	stopCh <- struct{}{}

	wg.Wait()
}

func TestWSStop(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-stop.yaml",
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
		&broadcast.Plugin{},
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
	t.Run("RPCWsStop", RPCWsMemoryStop("11114"))

	stopCh <- struct{}{}

	wg.Wait()
}

func RPCWsMemoryStop(port string) func(t *testing.T) {
	return func(t *testing.T) {
		da := websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: time.Second * 20,
		}

		connURL := url.URL{Scheme: "ws", Host: "localhost:" + port, Path: "/ws"}

		c, resp, err := da.Dial(connURL.String(), nil)
		assert.NotNil(t, resp)
		assert.Error(t, err)
		assert.Nil(t, c)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)    //nolint:staticcheck
		assert.Equal(t, resp.Header.Get("Stop"), "we-dont-like-you") //nolint:staticcheck
		if resp != nil && resp.Body != nil {                         //nolint:staticcheck
			_ = resp.Body.Close()
		}
	}
}

func TestWSAllow(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-allow.yaml",
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
		&broadcast.Plugin{},
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
	t.Run("RPCWsMemoryAllow", RPCWsPub("41278"))

	stopCh <- struct{}{}

	wg.Wait()
}

func TestWSAllow2(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-allow2.yaml",
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
		&broadcast.Plugin{},
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
	t.Run("RPCWsMemoryAllow", RPCWsPub("41270"))

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

	d, err := json.Marshal(messageWS("join", []byte("hello websockets"), "foo", "foo2"))
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

func RPCWsPubAsync(port string) func(t *testing.T) {
	return func(t *testing.T) {
		da := websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: time.Second * 20,
		}

		connURL := url.URL{Scheme: "ws", Host: "localhost:" + port, Path: "/ws"}

		c, resp, err := da.Dial(connURL.String(), nil)
		assert.NoError(t, err)

		defer func() {
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
		}()

		go func() {
			messagesToVerify := make([]string, 0, 10)
			messagesToVerify = append(messagesToVerify, `{"topic":"@join","payload":["foo","foo2"]}`)
			messagesToVerify = append(messagesToVerify, `{"topic":"foo","payload":"hello, PHP"}`)
			messagesToVerify = append(messagesToVerify, `{"topic":"@leave","payload":["foo"]}`)
			messagesToVerify = append(messagesToVerify, `{"topic":"foo2","payload":"hello, PHP2"}`)
			i := 0
			for {
				_, msg, err2 := c.ReadMessage()
				retMsg := utils.AsString(msg)
				assert.NoError(t, err2)
				assert.Equal(t, messagesToVerify[i], retMsg)
				i++
				if i == 3 {
					return
				}
			}
		}()

		time.Sleep(time.Second)

		d, err := json.Marshal(messageWS("join", []byte("hello websockets"), "foo", "foo2"))
		if err != nil {
			panic(err)
		}

		err = c.WriteMessage(websocket.BinaryMessage, d)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		publishAsync(t, "foo")

		time.Sleep(time.Second)

		// //// LEAVE foo /////////
		d, err = json.Marshal(messageWS("leave", []byte("hello websockets"), "foo"))
		if err != nil {
			panic(err)
		}

		err = c.WriteMessage(websocket.BinaryMessage, d)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		// TRY TO PUBLISH TO UNSUBSCRIBED TOPIC
		publishAsync(t, "foo")

		go func() {
			time.Sleep(time.Second * 5)
			publishAsync(t, "foo2")
		}()

		err = c.WriteControl(websocket.CloseMessage, nil, time.Time{})
		assert.NoError(t, err)
	}
}

func RPCWsPub(port string) func(t *testing.T) {
	return func(t *testing.T) {
		da := websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: time.Second * 20,
		}

		connURL := url.URL{Scheme: "ws", Host: "localhost:" + port, Path: "/ws"}

		c, resp, err := da.Dial(connURL.String(), nil)
		assert.NoError(t, err)

		defer func() {
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
		}()

		go func() {
			messagesToVerify := make([]string, 0, 10)
			messagesToVerify = append(messagesToVerify, `{"topic":"@join","payload":["foo","foo2"]}`)
			messagesToVerify = append(messagesToVerify, `{"topic":"foo","payload":"hello, PHP"}`)
			messagesToVerify = append(messagesToVerify, `{"topic":"@leave","payload":["foo"]}`)
			messagesToVerify = append(messagesToVerify, `{"topic":"foo2","payload":"hello, PHP2"}`)
			i := 0
			for {
				_, msg, err2 := c.ReadMessage()
				retMsg := utils.AsString(msg)
				assert.NoError(t, err2)
				assert.Equal(t, messagesToVerify[i], retMsg)
				i++
				if i == 3 {
					return
				}
			}
		}()

		time.Sleep(time.Second)

		d, err := json.Marshal(messageWS("join", []byte("hello websockets"), "foo", "foo2"))
		if err != nil {
			panic(err)
		}

		err = c.WriteMessage(websocket.BinaryMessage, d)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		publish("", "foo")

		time.Sleep(time.Second)

		// //// LEAVE foo /////////
		d, err = json.Marshal(messageWS("leave", []byte("hello websockets"), "foo"))
		if err != nil {
			panic(err)
		}

		err = c.WriteMessage(websocket.BinaryMessage, d)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		// TRY TO PUBLISH TO UNSUBSCRIBED TOPIC
		publish("", "foo")

		go func() {
			time.Sleep(time.Second * 5)
			publish2(t, "", "foo2")
		}()

		err = c.WriteControl(websocket.CloseMessage, nil, time.Time{})
		assert.NoError(t, err)
	}
}

func RPCWsDeny(port string) func(t *testing.T) {
	return func(t *testing.T) {
		da := websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: time.Second * 20,
		}

		connURL := url.URL{Scheme: "ws", Host: "localhost:" + port, Path: "/ws"}

		c, resp, err := da.Dial(connURL.String(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

		defer func() {
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
		}()

		d, err := json.Marshal(messageWS("join", []byte("hello websockets"), "foo", "foo2"))
		if err != nil {
			panic(err)
		}

		err = c.WriteMessage(websocket.BinaryMessage, d)
		assert.NoError(t, err)

		_, msg, err := c.ReadMessage()
		retMsg := utils.AsString(msg)
		assert.NoError(t, err)

		// subscription done
		assert.Equal(t, `{"topic":"#join","payload":["foo","foo2"]}`, retMsg)

		// //// LEAVE foo, foo2 /////////
		d, err = json.Marshal(messageWS("leave", []byte("hello websockets"), "foo"))
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

		err = c.WriteControl(websocket.CloseMessage, nil, time.Time{})
		assert.NoError(t, err)
	}
}

// ---------------------------------------------------------------------------------------------------

func publish(topics ...string) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	ret := &websocketsv1.Response{}
	err = client.Call("broadcast.Publish", makeMessage([]byte("hello, PHP"), topics...), ret)
	if err != nil {
		panic(err)
	}
}

func publishAsync(t *testing.T, topics ...string) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	ret := &websocketsv1.Response{}
	err = client.Call("broadcast.PublishAsync", makeMessage([]byte("hello, PHP"), topics...), ret)
	assert.NoError(t, err)
	assert.True(t, ret.Ok)
}

func publish2(t *testing.T, topics ...string) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	ret := &websocketsv1.Response{}
	err = client.Call("broadcast.Publish", makeMessage([]byte("hello, PHP2"), topics...), ret)
	assert.NoError(t, err)
	assert.True(t, ret.Ok)
}

func messageWS(command string, payload []byte, topics ...string) *websocketsv1.Message {
	return &websocketsv1.Message{
		Topics:  topics,
		Command: command,
		Payload: payload,
	}
}

func makeMessage(payload []byte, topics ...string) *websocketsv1.Request {
	m := &websocketsv1.Request{
		Messages: []*websocketsv1.Message{
			{
				Topics:  topics,
				Payload: payload,
			},
		},
	}

	return m
}
