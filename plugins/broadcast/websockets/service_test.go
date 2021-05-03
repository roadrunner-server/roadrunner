package websockets

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/broadcast/v2"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/env"
	rrhttp "github.com/spiral/roadrunner/service/http"
	"github.com/spiral/roadrunner/service/rpc"
	"github.com/stretchr/testify/assert"
)

type testCfg struct {
	http      string
	rpc       string
	ws        string
	broadcast string
	target    string
}

func (cfg *testCfg) Get(name string) service.Config {
	if name == rrhttp.ID {
		return &testCfg{target: cfg.http}
	}

	if name == ID {
		return &testCfg{target: cfg.ws}
	}

	if name == rpc.ID {
		return &testCfg{target: cfg.rpc}
	}

	if name == broadcast.ID {
		return &testCfg{target: cfg.broadcast}
	}

	return nil
}
func (cfg *testCfg) Unmarshal(out interface{}) error {
	return json.Unmarshal([]byte(cfg.target), out)
}

func readStr(m interface{}) string {
	return strings.TrimRight(string(m.([]byte)), "\n")
}

func Test_HttpService_Echo(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})

	assert.NoError(t, c.Init(&testCfg{
		http: `{
			"address": ":6041",
			"workers":{"command": "php tests/worker-ok.php", "pool.numWorkers": 1}
		}`,
	}))

	go func() { _ = c.Serve() }()
	time.Sleep(time.Millisecond * 3000)
	defer c.Stop()

	req, err := http.NewRequest("GET", "http://localhost:6041/", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		_ = r.Body.Close()
	}()

	b, _ := ioutil.ReadAll(r.Body)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, []byte(""), b)
}

func Test_HttpService_Echo400(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})

	assert.NoError(t, c.Init(&testCfg{
		http: `{
			"address": ":6040",
			"workers":{"command": "php tests/worker-stop.php", "pool.numWorkers": 1}
		}`,
	}))

	go func() { _ = c.Serve() }()
	time.Sleep(time.Millisecond * 3000)
	defer c.Stop()

	req, err := http.NewRequest("GET", "http://localhost:6040/", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		_ = r.Body.Close()
	}()

	assert.NoError(t, err)
	assert.Equal(t, 401, r.StatusCode)
}

func Test_Service_EnvPath(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(env.ID, &env.Service{})
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(broadcast.ID, &broadcast.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		http: `{
			"address": ":6029",
			"workers":{"command": "php tests/worker-ok.php", "pool.numWorkers": 1}
		}`,
		rpc:       `{"listen":"tcp://127.0.0.1:6002"}`,
		ws:        `{"path":"/ws"}`,
		broadcast: `{}`,
	}))

	go func() { _ = c.Serve() }()
	time.Sleep(time.Millisecond * 3000)
	defer c.Stop()

	req, err := http.NewRequest("GET", "http://localhost:6029/", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = r.Body.Close()
	}()

	b, _ := ioutil.ReadAll(r.Body)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, []byte("/ws"), b)
}

func Test_Service_Disabled(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(env.ID, &env.Service{})
	c.Register(broadcast.ID, &broadcast.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		ws:        `{"path":"/ws"}`,
		broadcast: `{}`,
	}))

	_, s := c.Get(ID)
	assert.Equal(t, service.StatusInactive, s)
}

func Test_Service_JoinTopic(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(env.ID, &env.Service{})
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(broadcast.ID, &broadcast.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		http: `{
			"address": ":6038",
			"workers":{"command": "php tests/worker-ok.php", "pool.numWorkers": 1}
		}`,
		rpc:       `{"listen":"tcp://127.0.0.1:6003"}`,
		ws:        `{"path":"/ws"}`,
		broadcast: `{}`,
	}))

	go func() { _ = c.Serve() }()
	time.Sleep(time.Millisecond * 1000)
	defer c.Stop()

	u := url.URL{Scheme: "ws", Host: "localhost:6038", Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()

	read := make(chan interface{})

	go func() {
		defer close(read)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			read <- message
		}
	}()

	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"topic":"join", "payload":["topic"]}`))
	assert.NoError(t, err)

	assert.Equal(t, `{"topic":"@join","payload":["topic"]}`, readStr(<-read))
}

func Test_Service_DenyJoin(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(env.ID, &env.Service{})
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(broadcast.ID, &broadcast.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		http: `{
			"address": ":6037",
			"workers":{"command": "php tests/worker-deny.php", "pool.numWorkers": 1}
		}`,
		rpc:       `{"listen":"tcp://127.0.0.1:6004"}`,
		ws:        `{"path":"/ws"}`,
		broadcast: `{}`,
	}))

	go func() { _ = c.Serve() }()
	time.Sleep(time.Millisecond * 1000)
	defer c.Stop()

	u := url.URL{Scheme: "ws", Host: "localhost:6037", Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()

	read := make(chan interface{})

	go func() {
		defer close(read)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				read <- err
				continue
			}
			read <- message
		}
	}()

	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"topic":"join", "payload":["topic"]}`))
	assert.NoError(t, err)

	assert.Equal(t, `{"topic":"#join","payload":["topic"]}`, readStr(<-read))
}

func Test_Service_DenyJoinServer(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(env.ID, &env.Service{})
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(broadcast.ID, &broadcast.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		http: `{
			"address": ":6037",
			"workers":{"command": "php tests/worker-stop.php", "pool.numWorkers": 1}
		}`,
		rpc:       `{"listen":"tcp://127.0.0.1:6005"}`,
		ws:        `{"path":"/ws"}`,
		broadcast: `{}`,
	}))

	go func() { _ = c.Serve() }()
	time.Sleep(time.Millisecond * 1000)
	defer c.Stop()

	u := url.URL{Scheme: "ws", Host: "localhost:6037", Path: "/ws"}

	_, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	assert.Error(t, err)
}

func Test_Service_EmptyTopics(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(env.ID, &env.Service{})
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(broadcast.ID, &broadcast.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		http: `{
			"address": ":6036",
			"workers":{"command": "php tests/worker-ok.php", "pool.numWorkers": 1}
		}`,
		rpc:       `{"listen":"tcp://127.0.0.1:6006"}`,
		ws:        `{"path":"/ws"}`,
		broadcast: `{}`,
	}))

	go func() { _ = c.Serve() }()
	time.Sleep(time.Millisecond * 1000)
	defer c.Stop()

	u := url.URL{Scheme: "ws", Host: "localhost:6036", Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()

	read := make(chan interface{})

	go func() {
		defer close(read)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				read <- err
				continue
			}
			read <- message
		}
	}()

	assert.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`{"topic":"join", "payload":[]}`)))

	assert.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`{"topic":"join", "payload":["a"]}`)))
	assert.Equal(t, `{"topic":"@join","payload":["a"]}`, readStr(<-read))

	assert.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`{"topic":"leave", "payload":[]}`)))

	assert.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`{"topic":"leave", "payload":["a"]}`)))
	assert.Equal(t, `{"topic":"@leave","payload":["a"]}`, readStr(<-read))

	// must be automatically closed during service stop
	assert.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`{"topic":"join", "payload":["a"]}`)))
	assert.Equal(t, `{"topic":"@join","payload":["a"]}`, readStr(<-read))
}

func Test_Service_BadTopics(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(env.ID, &env.Service{})
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(broadcast.ID, &broadcast.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		http: `{
			"address": ":6035",
			"workers":{"command": "php tests/worker-ok.php", "pool.numWorkers": 1}
		}`,
		rpc:       `{"listen":"tcp://127.0.0.1:6007"}`,
		ws:        `{"path":"/ws"}`,
		broadcast: `{}`,
	}))

	go func() { _ = c.Serve() }()
	time.Sleep(time.Millisecond * 1000)
	defer c.Stop()

	u := url.URL{Scheme: "ws", Host: "localhost:6035", Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()

	read := make(chan interface{})

	go func() {
		defer close(read)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				read <- err
				continue
			}
			read <- message
		}
	}()

	assert.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`{"topic":"join", "payload":"hello"}`)))
	assert.Error(t, (<-read).(error))
}

func Test_Service_BadTopicsLeave(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(env.ID, &env.Service{})
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(broadcast.ID, &broadcast.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		http: `{
			"address": ":6034",
			"workers":{"command": "php tests/worker-ok.php", "pool.numWorkers": 1}
		}`,
		rpc:       `{"listen":"tcp://127.0.0.1:6008"}`,
		ws:        `{"path":"/ws"}`,
		broadcast: `{}`,
	}))

	go func() { _ = c.Serve() }()
	time.Sleep(time.Millisecond * 1000)
	defer c.Stop()

	u := url.URL{Scheme: "ws", Host: "localhost:6034", Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()

	read := make(chan interface{})

	go func() {
		defer close(read)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				read <- err
				continue
			}
			read <- message
		}
	}()

	assert.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`{"topic":"leave", "payload":"hello"}`)))
	assert.Error(t, (<-read).(error))
}

func Test_Service_Events(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(env.ID, &env.Service{})
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(broadcast.ID, &broadcast.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		http: `{
			"address": ":6033",
			"workers":{"command": "php tests/worker-ok.php", "pool.numWorkers": 1}
		}`,
		rpc:       `{"listen":"tcp://127.0.0.1:6009"}`,
		ws:        `{"path":"/ws"}`,
		broadcast: `{}`,
	}))

	b, _ := c.Get(ID)
	br := b.(*Service)

	done := make(chan interface{})
	br.AddListener(func(event int, ctx interface{}) {
		if event == EventConnect {
			close(done)
		}
	})

	go func() { _ = c.Serve() }()
	time.Sleep(time.Millisecond * 1000)
	defer c.Stop()

	u := url.URL{Scheme: "ws", Host: "localhost:6033", Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()

	<-done

	read := make(chan interface{})

	go func() {
		defer close(read)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			read <- message
		}
	}()

	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"topic":"join", "payload":["topic"]}`))
	assert.NoError(t, err)

	assert.Equal(t, `{"topic":"@join","payload":["topic"]}`, readStr(<-read))
}

func Test_Service_Warmup(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(env.ID, &env.Service{})
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(broadcast.ID, &broadcast.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		http: `{
			"address": ":6033",
			"workers":{"command": "php tests/worker-ok.php", "pool.numWorkers": 1}
		}`,
		rpc:       `{"listen":"tcp://127.0.0.1:6009"}`,
		ws:        `{"path":"/ws"}`,
		broadcast: `{}`,
	}))

	rp, _ := c.Get(rpc.ID)

	b, _ := c.Get(ID)
	br := b.(*Service)

	done := make(chan interface{})
	br.AddListener(func(event int, ctx interface{}) {
		if event == EventConnect {
			close(done)
		}
	})

	go func() { _ = c.Serve() }()
	time.Sleep(time.Millisecond * 1000)
	defer c.Stop()

	client, err := rp.(*rpc.Service).Client()
	assert.NoError(t, err)

	var ok bool
	assert.NoError(t, client.Call("ws.SubscribePattern", "test", &ok))
	assert.True(t, ok)
	assert.NoError(t, client.Call("ws.Subscribe", "test", &ok))
	assert.True(t, ok)

	u := url.URL{Scheme: "ws", Host: "localhost:6033", Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()

	<-done

	read := make(chan interface{})

	go func() {
		defer close(read)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			read <- message
		}
	}()

	// not delivered
	assert.NoError(t, br.client.Publish(&broadcast.Message{Topic: "topic", Payload: []byte(`"hello"`)}))

	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"topic":"join", "payload":["topic"]}`))
	assert.NoError(t, err)

	assert.Equal(t, `{"topic":"@join","payload":["topic"]}`, readStr(<-read))

	assert.NoError(t, br.client.Publish(&broadcast.Message{Topic: "topic", Payload: []byte(`"hello"`)}))
	assert.Equal(t, `{"topic":"topic","payload":"hello"}`, readStr(<-read))
}

func Test_Service_Stop(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(env.ID, &env.Service{})
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(broadcast.ID, &broadcast.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		http: `{
			"address": ":6033",
			"workers":{"command": "php tests/worker-ok.php", "pool.numWorkers": 1}
		}`,
		rpc:       `{"listen":"tcp://127.0.0.1:6009"}`,
		ws:        `{"path":"/ws"}`,
		broadcast: `{}`,
	}))

	rp, _ := c.Get(rpc.ID)

	b, _ := c.Get(ID)
	br := b.(*Service)

	done := make(chan interface{})
	br.AddListener(func(event int, ctx interface{}) {
		if event == EventConnect {
			close(done)
		}
	})

	go func() { _ = c.Serve() }()
	time.Sleep(time.Millisecond * 1000)
	defer c.Stop()

	client, err := rp.(*rpc.Service).Client()
	assert.NoError(t, err)

	var ok bool
	assert.NoError(t, client.Call("ws.SubscribePattern", "test", &ok))
	assert.True(t, ok)
	assert.NoError(t, client.Call("ws.Subscribe", "test", &ok))
	assert.True(t, ok)

	u := url.URL{Scheme: "ws", Host: "localhost:6033", Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()

	<-done

	read := make(chan interface{})

	go func() {
		defer close(read)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			read <- message
		}
	}()

	// not delivered
	assert.NoError(t, br.client.Publish(&broadcast.Message{Topic: "topic", Payload: []byte(`"hello"`)}))

	br.Stop()

	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"topic":"join", "payload":["topic"]}`))
	assert.NoError(t, err)
}
