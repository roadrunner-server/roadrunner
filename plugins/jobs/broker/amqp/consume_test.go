package amqp

import (
	"fmt"
	"github.com/spiral/jobs/v2"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBroker_Consume_Job(t *testing.T) {
	b := &Broker{}
		_, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(pipe)
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
	assert.NoError(t, b.Consume(pipe, exec, func(id string, j *jobs.Job, err error) {}))

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	jid, perr := b.Push(pipe, &jobs.Job{
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

func TestBroker_ConsumeAfterStart_Job(t *testing.T) {
	b := &Broker{}
		_, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(pipe)
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

	exec := make(chan jobs.Handler, 1)
	assert.NoError(t, b.Consume(pipe, exec, func(id string, j *jobs.Job, err error) {}))

	<-ready

	jid, perr := b.Push(pipe, &jobs.Job{
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

func TestBroker_Consume_Delayed(t *testing.T) {
	b := &Broker{}
	_, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}

	err = b.Register(pipe)
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
	assert.NoError(t, b.Consume(pipe, exec, func(id string, j *jobs.Job, err error) {}))

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	start := time.Now()
	jid, perr := b.Push(pipe, &jobs.Job{
		Job:     "test",
		Payload: "body",
		Options: &jobs.Options{Delay: 1},
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

	elapsed := time.Since(start)
	assert.True(t, elapsed >= time.Second)
	assert.True(t, elapsed < 3*time.Second)
}

func TestBroker_Consume_Errored(t *testing.T) {
	b := &Broker{}
		_, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(pipe)
	if err != nil {
		t.Fatal(err)
	}

	ready := make(chan interface{})
	b.Listen(func(event int, ctx interface{}) {
		if event == jobs.EventBrokerReady {
			close(ready)
		}
	})

	errHandled := make(chan interface{})
	errHandler := func(id string, j *jobs.Job, err error) {
		assert.Equal(t, "job failed", err.Error())
		close(errHandled)
	}

	exec := make(chan jobs.Handler, 1)

	assert.NoError(t, b.Consume(pipe, exec, errHandler))

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	jid, perr := b.Push(pipe, &jobs.Job{Job: "test", Payload: "body", Options: &jobs.Options{}})
	assert.NotEqual(t, "", jid)
	assert.NoError(t, perr)

	waitJob := make(chan interface{})
	exec <- func(id string, j *jobs.Job) error {
		assert.Equal(t, jid, id)
		assert.Equal(t, "body", j.Payload)
		close(waitJob)
		return fmt.Errorf("job failed")
	}

	<-waitJob
	<-errHandled
}

func TestBroker_Consume_Errored_Attempts(t *testing.T) {
	b := &Broker{}
		_, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Register(pipe)
	if err != nil {
		t.Fatal(err)
	}

	ready := make(chan interface{})
	b.Listen(func(event int, ctx interface{}) {
		if event == jobs.EventBrokerReady {
			close(ready)
		}
	})

	attempts := 0
	errHandled := make(chan interface{})
	errHandler := func(id string, j *jobs.Job, err error) {
		assert.Equal(t, "job failed", err.Error())
		attempts++
		errHandled <- nil
	}

	exec := make(chan jobs.Handler, 1)

	assert.NoError(t, b.Consume(pipe, exec, errHandler))

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	jid, perr := b.Push(pipe, &jobs.Job{
		Job:     "test",
		Payload: "body",
		Options: &jobs.Options{Attempts: 3},
	})

	assert.NotEqual(t, "", jid)
	assert.NoError(t, perr)

	exec <- func(id string, j *jobs.Job) error {
		assert.Equal(t, jid, id)
		assert.Equal(t, "body", j.Payload)
		return fmt.Errorf("job failed")
	}

	<-errHandled
	<-errHandled
	<-errHandled
	assert.Equal(t, 3, attempts)
}
