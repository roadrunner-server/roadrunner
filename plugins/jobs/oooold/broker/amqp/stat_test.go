package amqp

import (
	"github.com/spiral/jobs/v2"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestBroker_Stat(t *testing.T) {
	b := &Broker{}
		_, err := b.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}
	b.Register(pipe)

	ready := make(chan interface{})
	b.Listen(func(event int, ctx interface{}) {
		if event == jobs.EventBrokerReady {
			close(ready)
		}
	})

	exec := make(chan jobs.Handler, 1)

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	jid, perr := b.Push(pipe, &jobs.Job{Job: "test", Payload: "body", Options: &jobs.Options{}})

	assert.NotEqual(t, "", jid)
	assert.NoError(t, perr)

	stat, err := b.Stat(pipe)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), stat.Queue)
	assert.Equal(t, int64(0), stat.Active)

	assert.NoError(t, b.Consume(pipe, exec, func(id string, j *jobs.Job, err error) {}))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	exec <- func(id string, j *jobs.Job) error {
		defer wg.Done()
		assert.Equal(t, jid, id)
		assert.Equal(t, "body", j.Payload)

		stat, err := b.Stat(pipe)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), stat.Active)

		return nil
	}

	wg.Wait()
	stat, err = b.Stat(pipe)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), stat.Queue)
	assert.Equal(t, int64(0), stat.Active)
}
