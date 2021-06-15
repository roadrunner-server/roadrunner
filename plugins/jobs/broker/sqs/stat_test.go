package sqs

import (
	"github.com/spiral/jobs/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBroker_Stat(t *testing.T) {
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

	go func() { assert.NoError(t, b.Serve()) }()
	defer b.Stop()

	<-ready

	jid, perr := b.Push(pipe, &jobs.Job{Job: "test", Payload: "body", Options: &jobs.Options{}})

	assert.NotEqual(t, "", jid)
	assert.NoError(t, perr)

	// unable to use approximated stats
	_, err = b.Stat(pipe)
	assert.NoError(t, err)

	assert.NoError(t, b.Consume(pipe, exec, func(id string, j *jobs.Job, err error) {}))

	waitJob := make(chan interface{})
	exec <- func(id string, j *jobs.Job) error {
		assert.Equal(t, jid, id)
		assert.Equal(t, "body", j.Payload)

		_, err := b.Stat(pipe)
		assert.NoError(t, err)

		close(waitJob)
		return nil
	}

	<-waitJob
	_, err = b.Stat(pipe)
	assert.NoError(t, err)
}
