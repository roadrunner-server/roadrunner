package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/client"
)

func Test_ExecuteChildWorkflow(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"WithChildWorkflow",
		"Hello World",
	)
	assert.NoError(t, err)

	var result string
	assert.NoError(t, w.Get(context.Background(), &result))
	assert.Equal(t, "Child: CHILD HELLO WORLD", result)
}

func Test_ExecuteChildStubWorkflow(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"WithChildStubWorkflow",
		"Hello World",
	)
	assert.NoError(t, err)

	var result string
	assert.NoError(t, w.Get(context.Background(), &result))
	assert.Equal(t, "Child: CHILD HELLO WORLD", result)
}

func Test_ExecuteChildStubWorkflow_02(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"ChildStubWorkflow",
		"Hello World",
	)
	assert.NoError(t, err)

	var result []string
	assert.NoError(t, w.Get(context.Background(), &result))
	assert.Equal(t, []string{"HELLO WORLD", "UNTYPED"}, result)
}

func Test_SignalChildViaStubWorkflow(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"SignalChildViaStubWorkflow",
	)
	assert.NoError(t, err)

	var result int
	assert.NoError(t, w.Get(context.Background(), &result))
	assert.Equal(t, 8, result)
}
