package ephemeral

import (
	"github.com/spiral/jobs/v2"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	pipe = &jobs.Pipeline{
		"broker": "local",
		"name":   "default",
	}
)

func TestBroker_Init(t *testing.T) {
	b := &Broker{}
	ok, err := b.Init()
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestBroker_StopNotStarted(t *testing.T) {
	b := &Broker{}
	_, err := b.Init()
	if err != nil {
		t.Fatal(err)
	}

	b.Stop()
}

func TestBroker_Register(t *testing.T) {
	b := &Broker{}
	b.Init()
	assert.NoError(t, b.Register(pipe))
}

func TestBroker_Register_Twice(t *testing.T) {
	b := &Broker{}
	_, err := b.Init()
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, b.Register(pipe))
	assert.Error(t, b.Register(pipe))
}

func TestBroker_Consume_Nil_BeforeServe(t *testing.T) {
	b := &Broker{}
	_, err := b.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(pipe)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, b.Consume(pipe, nil, nil))
}

func TestBroker_Consume_Undefined(t *testing.T) {
	b := &Broker{}
	_, err := b.Init()
	if err != nil {
		t.Fatal(err)
	}

	assert.Error(t, b.Consume(pipe, nil, nil))
}

func TestBroker_Consume_BeforeServe(t *testing.T) {
	b := &Broker{}
	_, err := b.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(pipe)
	if err != nil {
		t.Fatal(err)
	}

	exec := make(chan jobs.Handler)
	errf := func(id string, j *jobs.Job, err error) {}

	assert.NoError(t, b.Consume(pipe, exec, errf))
}

func TestBroker_Consume_Serve_Nil_Stop(t *testing.T) {
	b := &Broker{}
	_, err := b.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(pipe)
	if err != nil {
		t.Fatal(err)
	}

	err = b.Consume(pipe, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	wait := make(chan interface{})
	go func() {
		assert.NoError(t, b.Serve())
		close(wait)
	}()
	time.Sleep(time.Millisecond * 100)
	b.Stop()

	<-wait
}

func TestBroker_Consume_Serve_Stop(t *testing.T) {
	b := &Broker{}
	_, err := b.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(pipe)
	if err != nil {
		t.Fatal(err)
	}

	exec := make(chan jobs.Handler)
	errf := func(id string, j *jobs.Job, err error) {}

	err = b.Consume(pipe, exec, errf)
	if err != nil {
		t.Fatal(err)
	}

	wait := make(chan interface{})
	go func() {
		assert.NoError(t, b.Serve())
		close(wait)
	}()
	time.Sleep(time.Millisecond * 100)
	b.Stop()

	<-wait
}

func TestBroker_PushToNotRunning(t *testing.T) {
	b := &Broker{}
	_, err := b.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(pipe)
	if err != nil {
		t.Fatal(err)
	}

	_, err = b.Push(pipe, &jobs.Job{})
	assert.Error(t, err)
}

func TestBroker_StatNotRunning(t *testing.T) {
	b := &Broker{}
	_, err := b.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(pipe)
	if err != nil {
		t.Fatal(err)
	}

	_, err = b.Stat(pipe)
	assert.Error(t, err)
}

func TestBroker_PushToNotRegistered(t *testing.T) {
	b := &Broker{}
	_, err := b.Init()
	if err != nil {
		t.Fatal(err)
	}

	ready := make(chan interface{})
	b.Listen(func(event int, ctx interface{}) {
		if event == jobs.EventBrokerReady {
			close(ready)
		}
	})

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	_, err = b.Push(pipe, &jobs.Job{})
	assert.Error(t, err)
}

func TestBroker_StatNotRegistered(t *testing.T) {
	b := &Broker{}
	_, err := b.Init()
	if err != nil {
		t.Fatal(err)
	}

	ready := make(chan interface{})
	b.Listen(func(event int, ctx interface{}) {
		if event == jobs.EventBrokerReady {
			close(ready)
		}
	})

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	_, err = b.Stat(pipe)
	assert.Error(t, err)
}
