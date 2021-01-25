package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
)

func Test_SimpleWorkflowCancel(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"SimpleSignalledWorkflow")
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 500)
	err = s.Client().CancelWorkflow(context.Background(), w.GetID(), w.GetRunID())
	assert.NoError(t, err)

	var result interface{}
	assert.Error(t, w.Get(context.Background(), &result))

	we, err := s.Client().DescribeWorkflowExecution(context.Background(), w.GetID(), w.GetRunID())
	assert.NoError(t, err)

	assert.Equal(t, "Canceled", we.WorkflowExecutionInfo.Status.String())
}

func Test_CancellableWorkflowScope(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"CancelledScopeWorkflow",
		"Hello World",
	)
	assert.NoError(t, err)

	var result string
	assert.NoError(t, w.Get(context.Background(), &result))
	assert.Equal(t, "yes", result)

	s.AssertContainsEvent(t, w, func(event *history.HistoryEvent) bool {
		return event.EventType == enums.EVENT_TYPE_TIMER_CANCELED
	})

	s.AssertNotContainsEvent(t, w, func(event *history.HistoryEvent) bool {
		return event.EventType == enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED
	})
}

func Test_CancelledWorkflow(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"CancelledWorkflow",
		"Hello World",
	)
	assert.NoError(t, err)

	time.Sleep(time.Second)
	err = s.Client().CancelWorkflow(context.Background(), w.GetID(), w.GetRunID())
	assert.NoError(t, err)

	var result interface{}
	assert.NoError(t, w.Get(context.Background(), &result))
	assert.Equal(t, "CANCELLED", result)
}

func Test_CancelledWithCompensationWorkflow(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"CancelledWithCompensationWorkflow",
		"Hello World",
	)
	assert.NoError(t, err)

	time.Sleep(time.Second)
	err = s.Client().CancelWorkflow(context.Background(), w.GetID(), w.GetRunID())
	assert.NoError(t, err)

	var result interface{}
	assert.NoError(t, w.Get(context.Background(), &result))
	assert.Equal(t, "OK", result)

	e, err := s.Client().QueryWorkflow(context.Background(), w.GetID(), w.GetRunID(), "getStatus")
	assert.NoError(t, err)

	trace := make([]string, 0)
	assert.NoError(t, e.Get(&trace))
	assert.Equal(
		t,
		[]string{
			"yield",
			"rollback",
			"captured retry",
			"captured promise on cancelled",
			"START rollback",
			"WAIT ROLLBACK",
			"RESULT (ROLLBACK)", "DONE rollback",
			"COMPLETE rollback",
			"result: OK",
		},
		trace,
	)
}

func Test_CancelledNestedWorkflow(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"CancelledNestedWorkflow",
	)
	assert.NoError(t, err)

	time.Sleep(time.Second)

	err = s.Client().CancelWorkflow(context.Background(), w.GetID(), w.GetRunID())
	assert.NoError(t, err)

	var result interface{}
	assert.NoError(t, w.Get(context.Background(), &result))
	assert.Equal(t, "CANCELLED", result)

	e, err := s.Client().QueryWorkflow(context.Background(), w.GetID(), w.GetRunID(), "getStatus")
	assert.NoError(t, err)

	trace := make([]string, 0)
	assert.NoError(t, e.Get(&trace))
	assert.Equal(
		t,
		[]string{
			"begin",
			"first scope",
			"second scope",
			"close second scope",
			"close first scope",
			"second scope cancelled",
			"first scope cancelled",
			"close process",
		},
		trace,
	)
}

func Test_CancelledNSingleScopeWorkflow(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"CancelledSingleScopeWorkflow",
	)
	assert.NoError(t, err)

	time.Sleep(time.Second)

	err = s.Client().CancelWorkflow(context.Background(), w.GetID(), w.GetRunID())
	assert.NoError(t, err)

	var result interface{}
	assert.NoError(t, w.Get(context.Background(), &result))
	assert.Equal(t, "OK", result)

	e, err := s.Client().QueryWorkflow(context.Background(), w.GetID(), w.GetRunID(), "getStatus")
	assert.NoError(t, err)

	trace := make([]string, 0)
	assert.NoError(t, e.Get(&trace))
	assert.Equal(
		t,
		[]string{
			"start",
			"in scope",
			"on cancel",
			"captured in scope",
			"captured in process",
		},
		trace,
	)
}

func Test_CancelledMidflightWorkflow(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"CancelledMidflightWorkflow",
	)
	assert.NoError(t, err)

	var result interface{}
	assert.NoError(t, w.Get(context.Background(), &result))
	assert.Equal(t, "OK", result)

	e, err := s.Client().QueryWorkflow(context.Background(), w.GetID(), w.GetRunID(), "getStatus")
	assert.NoError(t, err)

	trace := make([]string, 0)
	assert.NoError(t, e.Get(&trace))
	assert.Equal(
		t,
		[]string{
			"start",
			"in scope",
			"on cancel",
			"done cancel",
		},
		trace,
	)

	s.AssertNotContainsEvent(t, w, func(event *history.HistoryEvent) bool {
		return event.EventType == enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED
	})
}

func Test_CancelSignalledChildWorkflow(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"CancelSignalledChildWorkflow",
	)
	assert.NoError(t, err)

	var result interface{}
	assert.NoError(t, w.Get(context.Background(), &result))
	assert.Equal(t, "cancelled ok", result)

	e, err := s.Client().QueryWorkflow(context.Background(), w.GetID(), w.GetRunID(), "getStatus")
	assert.NoError(t, err)

	trace := make([]string, 0)
	assert.NoError(t, e.Get(&trace))
	assert.Equal(
		t,
		[]string{
			"start",
			"child started",
			"child signalled",
			"scope cancelled",
			"process done",
		},
		trace,
	)

	s.AssertContainsEvent(t, w, func(event *history.HistoryEvent) bool {
		return event.EventType == enums.EVENT_TYPE_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_INITIATED
	})
}
