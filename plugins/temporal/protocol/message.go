package protocol

import (
	"time"

	"github.com/spiral/errors"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/activity"
	bindings "go.temporal.io/sdk/internalbindings"
	"go.temporal.io/sdk/workflow"
)

const (
	getWorkerInfoCommand = "GetWorkerInfo"

	invokeActivityCommand  = "InvokeActivity"
	startWorkflowCommand   = "StartWorkflow"
	invokeSignalCommand    = "InvokeSignal"
	invokeQueryCommand     = "InvokeQuery"
	destroyWorkflowCommand = "DestroyWorkflow"
	cancelWorkflowCommand  = "CancelWorkflow"
	getStackTraceCommand   = "StackTrace"

	executeActivityCommand           = "ExecuteActivity"
	executeChildWorkflowCommand      = "ExecuteChildWorkflow"
	getChildWorkflowExecutionCommand = "GetChildWorkflowExecution"

	newTimerCommand         = "NewTimer"
	sideEffectCommand       = "SideEffect"
	getVersionCommand       = "GetVersion"
	completeWorkflowCommand = "CompleteWorkflow"
	continueAsNewCommand    = "ContinueAsNew"

	signalExternalWorkflowCommand = "SignalExternalWorkflow"
	cancelExternalWorkflowCommand = "CancelExternalWorkflow"

	cancelCommand = "Cancel"
	panicCommand  = "Panic"
)

// GetWorkerInfo reads worker information.
type GetWorkerInfo struct {
}

// InvokeActivity invokes activity.
type InvokeActivity struct {
	// Name defines activity name.
	Name string `json:"name"`

	// Info contains execution context.
	Info activity.Info `json:"info"`

	// HeartbeatDetails indicates that the payload also contains last heartbeat details.
	HeartbeatDetails int `json:"heartbeatDetails,omitempty"`
}

// StartWorkflow sends worker command to start workflow.
type StartWorkflow struct {
	// Info to define workflow context.
	Info *workflow.Info `json:"info"`

	// LastCompletion contains offset of last completion results.
	LastCompletion int `json:"lastCompletion,omitempty"`
}

// InvokeSignal invokes signal with a set of arguments.
type InvokeSignal struct {
	// RunID workflow run id.
	RunID string `json:"runId"`

	// Name of the signal.
	Name string `json:"name"`
}

// InvokeQuery invokes query with a set of arguments.
type InvokeQuery struct {
	// RunID workflow run id.
	RunID string `json:"runId"`
	// Name of the query.
	Name string `json:"name"`
}

// CancelWorkflow asks worker to gracefully stop workflow, if possible (signal).
type CancelWorkflow struct {
	// RunID workflow run id.
	RunID string `json:"runId"`
}

// DestroyWorkflow asks worker to offload workflow from memory.
type DestroyWorkflow struct {
	// RunID workflow run id.
	RunID string `json:"runId"`
}

// GetStackTrace asks worker to offload workflow from memory.
type GetStackTrace struct {
	// RunID workflow run id.
	RunID string `json:"runId"`
}

// ExecuteActivity command by workflow worker.
type ExecuteActivity struct {
	// Name defines activity name.
	Name string `json:"name"`
	// Options to run activity.
	Options bindings.ExecuteActivityOptions `json:"options,omitempty"`
}

// ExecuteChildWorkflow executes child workflow.
type ExecuteChildWorkflow struct {
	// Name defines workflow name.
	Name string `json:"name"`
	// Options to run activity.
	Options bindings.WorkflowOptions `json:"options,omitempty"`
}

// GetChildWorkflowExecution returns the WorkflowID and RunId of child workflow.
type GetChildWorkflowExecution struct {
	// ID of child workflow command.
	ID uint64 `json:"id"`
}

// NewTimer starts new timer.
type NewTimer struct {
	// Milliseconds defines timer duration.
	Milliseconds int `json:"ms"`
}

// SideEffect to be recorded into the history.
type SideEffect struct{}

// GetVersion requests version marker.
type GetVersion struct {
	ChangeID     string `json:"changeID"`
	MinSupported int    `json:"minSupported"`
	MaxSupported int    `json:"maxSupported"`
}

// CompleteWorkflow sent by worker to complete workflow. Might include additional error as part of the payload.
type CompleteWorkflow struct{}

// ContinueAsNew restarts workflow with new running instance.
type ContinueAsNew struct {
	// Result defines workflow execution result.
	Name string `json:"name"`

	// Options for continued as new workflow.
	Options struct {
		TaskQueueName            string
		WorkflowExecutionTimeout time.Duration
		WorkflowRunTimeout       time.Duration
		WorkflowTaskTimeout      time.Duration
	} `json:"options"`
}

// SignalExternalWorkflow sends signal to external workflow.
type SignalExternalWorkflow struct {
	Namespace         string `json:"namespace"`
	WorkflowID        string `json:"workflowID"`
	RunID             string `json:"runID"`
	Signal            string `json:"signal"`
	ChildWorkflowOnly bool   `json:"childWorkflowOnly"`
}

// CancelExternalWorkflow canceller external workflow.
type CancelExternalWorkflow struct {
	Namespace  string `json:"namespace"`
	WorkflowID string `json:"workflowID"`
	RunID      string `json:"runID"`
}

// Cancel one or multiple internal promises (activities, local activities, timers, child workflows).
type Cancel struct {
	// CommandIDs to be cancelled.
	CommandIDs []uint64 `json:"ids"`
}

// Panic triggers panic in workflow process.
type Panic struct {
	// Message to include into the error.
	Message string `json:"message"`
}

// ActivityParams maps activity command to activity params.
func (cmd ExecuteActivity) ActivityParams(env bindings.WorkflowEnvironment, payloads *commonpb.Payloads) bindings.ExecuteActivityParams {
	params := bindings.ExecuteActivityParams{
		ExecuteActivityOptions: cmd.Options,
		ActivityType:           bindings.ActivityType{Name: cmd.Name},
		Input:                  payloads,
	}

	if params.TaskQueueName == "" {
		params.TaskQueueName = env.WorkflowInfo().TaskQueueName
	}

	return params
}

// WorkflowParams maps workflow command to workflow params.
func (cmd ExecuteChildWorkflow) WorkflowParams(env bindings.WorkflowEnvironment, payloads *commonpb.Payloads) bindings.ExecuteWorkflowParams {
	params := bindings.ExecuteWorkflowParams{
		WorkflowOptions: cmd.Options,
		WorkflowType:    &bindings.WorkflowType{Name: cmd.Name},
		Input:           payloads,
	}

	if params.TaskQueueName == "" {
		params.TaskQueueName = env.WorkflowInfo().TaskQueueName
	}

	return params
}

// ToDuration converts timer command to time.Duration.
func (cmd NewTimer) ToDuration() time.Duration {
	return time.Millisecond * time.Duration(cmd.Milliseconds)
}

// returns command name (only for the commands sent to the worker)
func commandName(cmd interface{}) (string, error) {
	switch cmd.(type) {
	case GetWorkerInfo, *GetWorkerInfo:
		return getWorkerInfoCommand, nil
	case StartWorkflow, *StartWorkflow:
		return startWorkflowCommand, nil
	case InvokeSignal, *InvokeSignal:
		return invokeSignalCommand, nil
	case InvokeQuery, *InvokeQuery:
		return invokeQueryCommand, nil
	case DestroyWorkflow, *DestroyWorkflow:
		return destroyWorkflowCommand, nil
	case CancelWorkflow, *CancelWorkflow:
		return cancelWorkflowCommand, nil
	case GetStackTrace, *GetStackTrace:
		return getStackTraceCommand, nil
	case InvokeActivity, *InvokeActivity:
		return invokeActivityCommand, nil
	case ExecuteActivity, *ExecuteActivity:
		return executeActivityCommand, nil
	case ExecuteChildWorkflow, *ExecuteChildWorkflow:
		return executeChildWorkflowCommand, nil
	case GetChildWorkflowExecution, *GetChildWorkflowExecution:
		return getChildWorkflowExecutionCommand, nil
	case NewTimer, *NewTimer:
		return newTimerCommand, nil
	case GetVersion, *GetVersion:
		return getVersionCommand, nil
	case SideEffect, *SideEffect:
		return sideEffectCommand, nil
	case CompleteWorkflow, *CompleteWorkflow:
		return completeWorkflowCommand, nil
	case ContinueAsNew, *ContinueAsNew:
		return continueAsNewCommand, nil
	case SignalExternalWorkflow, *SignalExternalWorkflow:
		return signalExternalWorkflowCommand, nil
	case CancelExternalWorkflow, *CancelExternalWorkflow:
		return cancelExternalWorkflowCommand, nil
	case Cancel, *Cancel:
		return cancelCommand, nil
	case Panic, *Panic:
		return panicCommand, nil
	default:
		return "", errors.E(errors.Op("commandName"), "undefined command type", cmd)
	}
}

// reads command from binary payload
func initCommand(name string) (interface{}, error) {
	switch name {
	case getWorkerInfoCommand:
		return &GetWorkerInfo{}, nil

	case startWorkflowCommand:
		return &StartWorkflow{}, nil

	case invokeSignalCommand:
		return &InvokeSignal{}, nil

	case invokeQueryCommand:
		return &InvokeQuery{}, nil

	case destroyWorkflowCommand:
		return &DestroyWorkflow{}, nil

	case cancelWorkflowCommand:
		return &CancelWorkflow{}, nil

	case getStackTraceCommand:
		return &GetStackTrace{}, nil

	case invokeActivityCommand:
		return &InvokeActivity{}, nil

	case executeActivityCommand:
		return &ExecuteActivity{}, nil

	case executeChildWorkflowCommand:
		return &ExecuteChildWorkflow{}, nil

	case getChildWorkflowExecutionCommand:
		return &GetChildWorkflowExecution{}, nil

	case newTimerCommand:
		return &NewTimer{}, nil

	case getVersionCommand:
		return &GetVersion{}, nil

	case sideEffectCommand:
		return &SideEffect{}, nil

	case completeWorkflowCommand:
		return &CompleteWorkflow{}, nil

	case continueAsNewCommand:
		return &ContinueAsNew{}, nil

	case signalExternalWorkflowCommand:
		return &SignalExternalWorkflow{}, nil

	case cancelExternalWorkflowCommand:
		return &CancelExternalWorkflow{}, nil

	case cancelCommand:
		return &Cancel{}, nil

	case panicCommand:
		return &Panic{}, nil

	default:
		return nil, errors.E(errors.Op("initCommand"), "undefined command type", name)
	}
}
