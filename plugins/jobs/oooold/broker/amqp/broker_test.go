package amqp

import (
	"github.com/spiral/jobs/v2"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	pipe = &jobs.Pipeline{
		"broker":   "amqp",
		"name":     "default",
		"queue":    "rr-queue",
		"exchange": "rr-exchange",
		"prefetch": 1,
	}

	cfg = &Config{
		Addr: "amqp://guest:guest@localhost:5672/",
	}
)

var (
	fanoutPipe = &jobs.Pipeline{
		"broker":   "amqp",
		"name":     "fanout",
		"queue":    "fanout-queue",
		"exchange": "fanout-exchange",
		"exchange-type": "fanout",
		"prefetch": 1,
	}

	fanoutCfg = &Config{
		Addr: "amqp://guest:guest@localhost:5672/",
	}
)

func TestBroker_Init(t *testing.T) {
	b := &Broker{}
	ok, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestBroker_StopNotStarted(t *testing.T) {
	b := &Broker{}
	_, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}

	b.Stop()
}

func TestBroker_Register(t *testing.T) {
	b := &Broker{}
	_, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, b.Register(pipe))
}

func TestBroker_Register_Twice(t *testing.T) {
	b := &Broker{}
	_, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, b.Register(pipe))
	assert.Error(t, b.Register(pipe))
}

func TestBroker_Consume_Nil_BeforeServe(t *testing.T) {
	b := &Broker{}
	_, err := b.Init(cfg)
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
	_, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}

	assert.Error(t, b.Consume(pipe, nil, nil))
}

func TestBroker_Consume_BeforeServe(t *testing.T) {
	b := &Broker{}
	_, err := b.Init(cfg)
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

func TestBroker_Consume_BadPipeline(t *testing.T) {
	b := &Broker{}
	_, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}
	assert.Error(t, b.Register(&jobs.Pipeline{
		"broker":   "amqp",
		"name":     "default",
		"exchange": "rr-exchange",
		"prefetch": 1,
	}))
}

func TestBroker_Consume_Serve_Nil_Stop(t *testing.T) {
	b := &Broker{}
	_, err := b.Init(cfg)
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

func TestBroker_Consume_CantStart(t *testing.T) {
	b := &Broker{}
	_, err := b.Init(&Config{
		Addr: "amqp://guest:guest@localhost:15672/",
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Error(t, b.Serve())
}

func TestBroker_Consume_Serve_Stop(t *testing.T) {
	b := &Broker{}
	_, err := b.Init(cfg)
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
		t.Fatal()
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
	_, err := b.Init(cfg)
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
	_, err := b.Init(cfg)
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
	_, err := b.Init(cfg)
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
	_, err := b.Init(cfg)
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

func TestBroker_Queue_RoutingKey(t *testing.T) {
	pipeWithKey := pipe.With("routing-key", "rr-exchange-routing-key")

	assert.Equal(t, pipeWithKey.String("routing-key", ""), "rr-exchange-routing-key")
}

func TestBroker_Register_With_RoutingKey(t *testing.T) {
	b := &Broker{}
	_, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}

	pipeWithKey := pipe.With("routing-key", "rr-exchange-routing-key")

	assert.NoError(t, b.Register(&pipeWithKey))
}

func TestBroker_Consume_With_RoutingKey(t *testing.T) {
	b := &Broker{}
	_, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}

	pipeWithKey := pipe.With("routing-key", "rr-exchange-routing-key")

	err = b.Register(&pipeWithKey)
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
	assert.NoError(t, b.Consume(&pipeWithKey, exec, func(id string, j *jobs.Job, err error) {}))

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	jid, perr := b.Push(&pipeWithKey, &jobs.Job{
		Job:     "test",
		Payload: "body",
		Options: &jobs.Options{},
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

func TestBroker_Queue_ExchangeType(t *testing.T) {
	pipeWithKey := pipe.With("exchange-type", "direct")

	assert.Equal(t, pipeWithKey.String("exchange-type", ""), "direct")
}

func TestBroker_Register_With_ExchangeType(t *testing.T) {
	b := &Broker{}
	_, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}

	pipeWithKey := pipe.With("exchange-type", "fanout")

	assert.NoError(t, b.Register(&pipeWithKey))
}

func TestBroker_Register_With_WrongExchangeType(t *testing.T) {
	b := &Broker{}
	_, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}

	pipeWithKey := pipe.With("exchange-type", "xxx")

	assert.Error(t, b.Register(&pipeWithKey))
}

func TestBroker_Consume_With_ExchangeType(t *testing.T) {
	b := &Broker{}
	_, err := b.Init(fanoutCfg)
	if err != nil {
		t.Fatal(err)
	}

	pipeWithKey := fanoutPipe.With("exchange-type", "fanout")

	err = b.Register(&pipeWithKey)
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
	assert.NoError(t, b.Consume(&pipeWithKey, exec, func(id string, j *jobs.Job, err error) {}))

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	jid, perr := b.Push(&pipeWithKey, &jobs.Job{
		Job:     "test",
		Payload: "body",
		Options: &jobs.Options{},
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
