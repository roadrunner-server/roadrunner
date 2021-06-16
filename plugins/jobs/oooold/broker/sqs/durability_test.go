package sqs

import (
	"github.com/spiral/jobs/v2"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

var (
	proxyCfg = &Config{
		Key:      "api-key",
		Secret:   "api-secret",
		Region:   "us-west-1",
		Endpoint: "http://localhost:9325",
		Timeout:  1,
	}

	proxy = &tcpProxy{
		listen:   "localhost:9325",
		upstream: "localhost:9324",
		accept:   true,
	}

	proxyPipe = &jobs.Pipeline{
		"broker":       "sqs",
		"name":         "default",
		"queue":        "test",
		"lockReserved": 1,
		"declare": map[string]interface{}{
			"MessageRetentionPeriod": 86400,
		},
	}
)

type tcpProxy struct {
	listen   string
	upstream string
	mu       sync.Mutex
	accept   bool
	conn     []net.Conn
}

func (p *tcpProxy) serve() {
	l, err := net.Listen("tcp", p.listen)
	if err != nil {
		panic(err)
	}

	for {
		in, err := l.Accept()
		if err != nil {
			panic(err)
		}

		if !p.accepting() {
			in.Close()
		}

		up, err := net.Dial("tcp", p.upstream)
		if err != nil {
			panic(err)
		}

		go io.Copy(in, up)
		go io.Copy(up, in)

		p.mu.Lock()
		p.conn = append(p.conn, in, up)
		p.mu.Unlock()
	}
}

// wait for specific number of connections
func (p *tcpProxy) waitConn(count int) *tcpProxy {
	p.mu.Lock()
	p.accept = true
	p.mu.Unlock()

	for {
		p.mu.Lock()
		current := len(p.conn)
		p.mu.Unlock()

		if current >= count*2 {
			break
		}

		time.Sleep(time.Millisecond)
	}

	return p
}

func (p *tcpProxy) reset(accept bool) int {
	p.mu.Lock()
	p.accept = accept
	defer p.mu.Unlock()

	count := 0
	for _, conn := range p.conn {
		conn.Close()
		count++
	}

	p.conn = nil
	return count / 2
}

func (p *tcpProxy) accepting() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.accept
}

func init() {
	go proxy.serve()
}

func TestBroker_Durability_Base(t *testing.T) {
	defer proxy.reset(true)

	b := &Broker{}
	_, err := b.Init(proxyCfg)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(proxyPipe)
	if err != nil {
		t.Fatal(err)
	}

	ready := make(chan interface{})
	b.Listen(func(event int, ctx interface{}) {
		if event == jobs.EventBrokerReady {
			close(ready)
		}
	})

	exec := make(chan jobs.Handler, 1)
	assert.NoError(t, b.Consume(proxyPipe, exec, func(id string, j *jobs.Job, err error) {}))

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	// expect 2 connections
	proxy.waitConn(1)

	jid, perr := b.Push(proxyPipe, &jobs.Job{
		Job:     "test",
		Payload: "body",
		Options: &jobs.Options{Timeout: 2},
	})

	assert.NotEqual(t, "", jid)
	assert.NoError(t, perr)

	waitJob := make(chan interface{})
	exec <- func(id string, j *jobs.Job) error {
		assert.Equal(t, jid, id)
		assert.Equal(t, "body", j.Payload)
		close(waitJob)
		return nil
	}

	<-waitJob
}

func TestBroker_Durability_Consume(t *testing.T) {
	defer proxy.reset(true)

	b := &Broker{}
	_, err := b.Init(proxyCfg)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(proxyPipe)
	if err != nil {
		t.Fatal(err)
	}
	ready := make(chan interface{})
	b.Listen(func(event int, ctx interface{}) {
		if event == jobs.EventBrokerReady {
			close(ready)
		}
	})

	exec := make(chan jobs.Handler, 1)
	assert.NoError(t, b.Consume(proxyPipe, exec, func(id string, j *jobs.Job, err error) {}))

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	proxy.waitConn(1).reset(false)

	jid, perr := b.Push(proxyPipe, &jobs.Job{
		Job:     "test",
		Payload: "body",
		Options: &jobs.Options{},
	})

	assert.Error(t, perr)

	// restore
	proxy.waitConn(1)

	jid, perr = b.Push(proxyPipe, &jobs.Job{
		Job:     "test",
		Payload: "body",
		Options: &jobs.Options{Timeout: 2},
	})

	assert.NotEqual(t, "", jid)
	assert.NoError(t, perr)

	mu := sync.Mutex{}
	done := make(map[string]bool)
	exec <- func(id string, j *jobs.Job) error {
		mu.Lock()
		defer mu.Unlock()
		done[id] = true

		assert.Equal(t, jid, id)
		assert.Equal(t, "body", j.Payload)

		return nil
	}

	for {
		mu.Lock()
		num := len(done)
		mu.Unlock()

		if num >= 1 {
			break
		}
	}
}

func TestBroker_Durability_Consume_LongTimeout(t *testing.T) {
	defer proxy.reset(true)

	b := &Broker{}
	_, err := b.Init(proxyCfg)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(proxyPipe)
	if err != nil {
		t.Fatal(err)
	}
	ready := make(chan interface{})
	b.Listen(func(event int, ctx interface{}) {
		if event == jobs.EventBrokerReady {
			close(ready)
		}
	})

	exec := make(chan jobs.Handler, 1)
	assert.NoError(t, b.Consume(proxyPipe, exec, func(id string, j *jobs.Job, err error) {}))

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	proxy.waitConn(1).reset(false)

	jid, perr := b.Push(proxyPipe, &jobs.Job{
		Job:     "test",
		Payload: "body",
		Options: &jobs.Options{},
	})

	assert.Error(t, perr)

	// restore
	time.Sleep(3 * time.Second)
	proxy.waitConn(1)

	jid, perr = b.Push(proxyPipe, &jobs.Job{
		Job:     "test",
		Payload: "body",
		Options: &jobs.Options{Timeout: 2},
	})

	assert.NotEqual(t, "", jid)
	assert.NoError(t, perr)

	mu := sync.Mutex{}
	done := make(map[string]bool)
	exec <- func(id string, j *jobs.Job) error {
		mu.Lock()
		defer mu.Unlock()
		done[id] = true

		assert.Equal(t, jid, id)
		assert.Equal(t, "body", j.Payload)

		return nil
	}

	for {
		mu.Lock()
		num := len(done)
		mu.Unlock()

		if num >= 1 {
			break
		}
	}
}

func TestBroker_Durability_Consume2(t *testing.T) {
	defer proxy.reset(true)

	b := &Broker{}
	_, err := b.Init(proxyCfg)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(proxyPipe)
	if err != nil {
		t.Fatal(err)
	}
	ready := make(chan interface{})
	b.Listen(func(event int, ctx interface{}) {
		if event == jobs.EventBrokerReady {
			close(ready)
		}
	})

	exec := make(chan jobs.Handler, 1)
	assert.NoError(t, b.Consume(proxyPipe, exec, func(id string, j *jobs.Job, err error) {}))

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	proxy.waitConn(1).reset(false)

	jid, perr := b.Push(proxyPipe, &jobs.Job{
		Job:     "test",
		Payload: "body",
		Options: &jobs.Options{},
	})

	assert.Error(t, perr)

	// restore
	proxy.waitConn(2)

	jid, perr = b.Push(proxyPipe, &jobs.Job{
		Job:     "test",
		Payload: "body",
		Options: &jobs.Options{Timeout: 2},
	})

	assert.NotEqual(t, "", jid)
	assert.NoError(t, perr)

	st, serr := b.Stat(proxyPipe)
	assert.NoError(t, serr)
	assert.Equal(t, int64(1), st.Queue+st.Active)

	proxy.reset(false)

	_, serr = b.Stat(proxyPipe)
	assert.Error(t, serr)

	proxy.reset(true)

	_, serr = b.Stat(proxyPipe)
	assert.NoError(t, serr)

	mu := sync.Mutex{}
	done := make(map[string]bool)
	exec <- func(id string, j *jobs.Job) error {
		mu.Lock()
		defer mu.Unlock()
		done[id] = true

		assert.Equal(t, jid, id)
		assert.Equal(t, "body", j.Payload)

		return nil
	}

	for {
		mu.Lock()
		num := len(done)
		mu.Unlock()

		if num >= 1 {
			break
		}
	}
}

func TestBroker_Durability_Consume3(t *testing.T) {
	defer proxy.reset(true)

	b := &Broker{}
	_, err := b.Init(proxyCfg)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(proxyPipe)
	if err != nil {
		t.Fatal(err)
	}
	ready := make(chan interface{})
	b.Listen(func(event int, ctx interface{}) {
		if event == jobs.EventBrokerReady {
			close(ready)
		}
	})

	exec := make(chan jobs.Handler, 1)
	assert.NoError(t, b.Consume(proxyPipe, exec, func(id string, j *jobs.Job, err error) {}))

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	proxy.waitConn(1)

	jid, perr := b.Push(proxyPipe, &jobs.Job{
		Job:     "test",
		Payload: "body",
		Options: &jobs.Options{Timeout: 2},
	})

	assert.NotEqual(t, "", jid)
	assert.NoError(t, perr)

	st, serr := b.Stat(proxyPipe)
	assert.NoError(t, serr)
	assert.Equal(t, int64(1), st.Queue+st.Active)

	mu := sync.Mutex{}
	done := make(map[string]bool)
	exec <- func(id string, j *jobs.Job) error {
		mu.Lock()
		defer mu.Unlock()
		done[id] = true

		assert.Equal(t, jid, id)
		assert.Equal(t, "body", j.Payload)

		return nil
	}

	for {
		mu.Lock()
		num := len(done)
		mu.Unlock()

		if num >= 1 {
			break
		}
	}
}

func TestBroker_Durability_Consume4(t *testing.T) {
	defer proxy.reset(true)

	b := &Broker{}
	_, err := b.Init(proxyCfg)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(proxyPipe)
	if err != nil {
		t.Fatal(err)
	}
	ready := make(chan interface{})
	b.Listen(func(event int, ctx interface{}) {
		if event == jobs.EventBrokerReady {
			close(ready)
		}
	})

	exec := make(chan jobs.Handler, 1)
	assert.NoError(t, b.Consume(proxyPipe, exec, func(id string, j *jobs.Job, err error) {}))

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	proxy.waitConn(1)

	_, err = b.Push(proxyPipe, &jobs.Job{
		Job:     "test",
		Payload: "kill",
		Options: &jobs.Options{Timeout: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.Push(proxyPipe, &jobs.Job{
		Job:     "test",
		Payload: "body",
		Options: &jobs.Options{Timeout: 2},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = b.Push(proxyPipe, &jobs.Job{
		Job:     "test",
		Payload: "body",
		Options: &jobs.Options{Timeout: 2},
	})
	if err != nil {
		t.Fatal(err)
	}

	st, serr := b.Stat(proxyPipe)
	assert.NoError(t, serr)
	assert.Equal(t, int64(3), st.Queue+st.Active)

	mu := sync.Mutex{}
	done := make(map[string]bool)
	exec <- func(id string, j *jobs.Job) error {
		mu.Lock()
		defer mu.Unlock()
		done[id] = true

		if j.Payload == "kill" {
			proxy.reset(true)
		}

		return nil
	}

	for {
		mu.Lock()
		num := len(done)
		mu.Unlock()

		if num >= 3 {
			break
		}
	}
}

func TestBroker_Durability_StopDead(t *testing.T) {
	defer proxy.reset(true)

	b := &Broker{}
	_, err := b.Init(proxyCfg)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(proxyPipe)
	if err != nil {
		t.Fatal(err)
	}
	ready := make(chan interface{})
	b.Listen(func(event int, ctx interface{}) {
		if event == jobs.EventBrokerReady {
			close(ready)
		}
	})

	exec := make(chan jobs.Handler, 1)
	assert.NoError(t, b.Consume(proxyPipe, exec, func(id string, j *jobs.Job, err error) {}))

	go func() { assert.NoError(t, b.Serve()) }()

	<-ready

	proxy.waitConn(1).reset(false)

	b.Stop()
}
