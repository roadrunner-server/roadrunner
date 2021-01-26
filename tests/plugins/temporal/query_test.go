package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/client"
)

func Test_ListQueries(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"QueryWorkflow",
		"Hello World",
	)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 500)

	v, err := s.Client().QueryWorkflow(context.Background(), w.GetID(), w.GetRunID(), "error", -1)
	assert.Nil(t, v)
	assert.Error(t, err)

	assert.Contains(t, err.Error(), "KnownQueryTypes=[get]")

	var r int
	assert.NoError(t, w.Get(context.Background(), &r))
	assert.Equal(t, 0, r)
}

func Test_GetQuery(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"QueryWorkflow",
		"Hello World",
	)
	assert.NoError(t, err)

	err = s.Client().SignalWorkflow(context.Background(), w.GetID(), w.GetRunID(), "add", 88)
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 500)

	v, err := s.Client().QueryWorkflow(context.Background(), w.GetID(), w.GetRunID(), "get", nil)
	assert.NoError(t, err)

	var r int
	assert.NoError(t, v.Get(&r))
	assert.Equal(t, 88, r)

	assert.NoError(t, w.Get(context.Background(), &r))
	assert.Equal(t, 88, r)
}
