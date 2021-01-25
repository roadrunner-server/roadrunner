package tests

import (
	"context"
	"testing"
	"time"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
)

func Test_SignalsWithoutSignals(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"SimpleSignalledWorkflow",
		"Hello World",
	)
	assert.NoError(t, err)

	var result int
	assert.NoError(t, w.Get(context.Background(), &result))
	assert.Equal(t, 0, result)
}

func Test_SendSignalDuringTimer(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().SignalWithStartWorkflow(
		context.Background(),
		"signalled-"+uuid.New(),
		"add",
		10,
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"SimpleSignalledWorkflow",
	)
	assert.NoError(t, err)

	err = s.Client().SignalWorkflow(context.Background(), w.GetID(), w.GetRunID(), "add", -1)
	assert.NoError(t, err)

	var result int
	assert.NoError(t, w.Get(context.Background(), &result))
	assert.Equal(t, 9, result)

	s.AssertContainsEvent(t, w, func(event *history.HistoryEvent) bool {
		if event.EventType == enums.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED {
			attr := event.Attributes.(*history.HistoryEvent_WorkflowExecutionSignaledEventAttributes)
			return attr.WorkflowExecutionSignaledEventAttributes.SignalName == "add"
		}

		return false
	})
}

func Test_SendSignalBeforeCompletingWorkflow(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"SimpleSignalledWorkflowWithSleep",
	)
	assert.NoError(t, err)

	// should be around sleep(1) call
	time.Sleep(time.Second + time.Millisecond*200)

	err = s.Client().SignalWorkflow(context.Background(), w.GetID(), w.GetRunID(), "add", -1)
	assert.NoError(t, err)

	var result int
	assert.NoError(t, w.Get(context.Background(), &result))
	assert.Equal(t, -1, result)

	s.AssertContainsEvent(t, w, func(event *history.HistoryEvent) bool {
		if event.EventType == enums.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED {
			attr := event.Attributes.(*history.HistoryEvent_WorkflowExecutionSignaledEventAttributes)
			return attr.WorkflowExecutionSignaledEventAttributes.SignalName == "add"
		}

		return false
	})
}

func Test_RuntimeSignal(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().SignalWithStartWorkflow(
		context.Background(),
		"signalled-"+uuid.New(),
		"add",
		-1,
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"RuntimeSignalWorkflow",
	)
	assert.NoError(t, err)

	var result int
	assert.NoError(t, w.Get(context.Background(), &result))
	assert.Equal(t, -1, result)

	s.AssertContainsEvent(t, w, func(event *history.HistoryEvent) bool {
		if event.EventType == enums.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED {
			attr := event.Attributes.(*history.HistoryEvent_WorkflowExecutionSignaledEventAttributes)
			return attr.WorkflowExecutionSignaledEventAttributes.SignalName == "add"
		}

		return false
	})
}

func Test_SignalSteps(t *testing.T) {
	s := NewTestServer()
	defer s.MustClose()

	w, err := s.Client().ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: "default",
		},
		"WorkflowWithSignalledSteps",
	)
	assert.NoError(t, err)

	err = s.Client().SignalWorkflow(context.Background(), w.GetID(), w.GetRunID(), "begin", true)
	assert.NoError(t, err)

	err = s.Client().SignalWorkflow(context.Background(), w.GetID(), w.GetRunID(), "next1", true)
	assert.NoError(t, err)

	v, err := s.Client().QueryWorkflow(context.Background(), w.GetID(), w.GetRunID(), "value", nil)
	assert.NoError(t, err)

	var r int
	assert.NoError(t, v.Get(&r))
	assert.Equal(t, 2, r)

	err = s.Client().SignalWorkflow(context.Background(), w.GetID(), w.GetRunID(), "next2", true)
	assert.NoError(t, err)

	v, err = s.Client().QueryWorkflow(context.Background(), w.GetID(), w.GetRunID(), "value", nil)
	assert.NoError(t, err)

	assert.NoError(t, v.Get(&r))
	assert.Equal(t, 3, r)

	var result int
	assert.NoError(t, w.Get(context.Background(), &result))

	// 3 ticks
	assert.Equal(t, 3, result)
}
