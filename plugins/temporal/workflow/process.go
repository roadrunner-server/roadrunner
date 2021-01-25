package workflow

import (
	"strconv"
	"sync/atomic"
	"time"

	"github.com/spiral/errors"
	rrt "github.com/spiral/roadrunner/v2/plugins/temporal/protocol"
	commonpb "go.temporal.io/api/common/v1"
	bindings "go.temporal.io/sdk/internalbindings"
	"go.temporal.io/sdk/workflow"
)

// wraps single workflow process
type workflowProcess struct {
	codec     rrt.Codec
	pool      workflowPool
	env       bindings.WorkflowEnvironment
	header    *commonpb.Header
	mq        *messageQueue
	ids       *idRegistry
	seqID     uint64
	runID     string
	pipeline  []rrt.Message
	callbacks []func() error
	canceller *canceller
	inLoop    bool
}

// Execute workflow, bootstraps process.
func (wf *workflowProcess) Execute(env bindings.WorkflowEnvironment, header *commonpb.Header, input *commonpb.Payloads) {
	wf.env = env
	wf.header = header
	wf.seqID = 0
	wf.runID = env.WorkflowInfo().WorkflowExecution.RunID
	wf.canceller = &canceller{}

	// sequenceID shared for all worker workflows
	wf.mq = newMessageQueue(wf.pool.SeqID)
	wf.ids = newIDRegistry()

	env.RegisterCancelHandler(wf.handleCancel)
	env.RegisterSignalHandler(wf.handleSignal)
	env.RegisterQueryHandler(wf.handleQuery)

	var (
		lastCompletion       = bindings.GetLastCompletionResult(env)
		lastCompletionOffset = 0
	)

	if lastCompletion != nil && len(lastCompletion.Payloads) != 0 {
		if input == nil {
			input = &commonpb.Payloads{Payloads: []*commonpb.Payload{}}
		}

		input.Payloads = append(input.Payloads, lastCompletion.Payloads...)
		lastCompletionOffset = len(lastCompletion.Payloads)
	}

	_, err := wf.mq.pushCommand(
		rrt.StartWorkflow{
			Info:           env.WorkflowInfo(),
			LastCompletion: lastCompletionOffset,
		},
		input,
	)

	if err != nil {
		panic(err)
	}
}

// OnWorkflowTaskStarted handles single workflow tick and batch of pipeline from temporal server.
func (wf *workflowProcess) OnWorkflowTaskStarted() {
	wf.inLoop = true
	defer func() { wf.inLoop = false }()

	var err error
	for _, callback := range wf.callbacks {
		err = callback()
		if err != nil {
			panic(err)
		}
	}
	wf.callbacks = nil

	if err := wf.flushQueue(); err != nil {
		panic(err)
	}

	for len(wf.pipeline) > 0 {
		msg := wf.pipeline[0]
		wf.pipeline = wf.pipeline[1:]

		if msg.IsCommand() {
			err = wf.handleMessage(msg)
		}

		if err != nil {
			panic(err)
		}
	}
}

// StackTrace renders workflow stack trace.
func (wf *workflowProcess) StackTrace() string {
	result, err := wf.runCommand(
		rrt.GetStackTrace{
			RunID: wf.env.WorkflowInfo().WorkflowExecution.RunID,
		},
		nil,
	)

	if err != nil {
		return err.Error()
	}

	var stacktrace string
	err = wf.env.GetDataConverter().FromPayload(result.Payloads.Payloads[0], &stacktrace)
	if err != nil {
		return err.Error()
	}

	return stacktrace
}

// Close the workflow.
func (wf *workflowProcess) Close() {
	// TODO: properly handle errors
	// panic(err)

	_, err := wf.mq.pushCommand(
		rrt.DestroyWorkflow{RunID: wf.env.WorkflowInfo().WorkflowExecution.RunID},
		nil,
	)

	if err != nil {
		panic(err)
	}

	_, _ = wf.discardQueue()
}

// execution context.
func (wf *workflowProcess) getContext() rrt.Context {
	return rrt.Context{
		TaskQueue: wf.env.WorkflowInfo().TaskQueueName,
		TickTime:  wf.env.Now().Format(time.RFC3339),
		Replay:    wf.env.IsReplaying(),
	}
}

// schedule cancel command
func (wf *workflowProcess) handleCancel() {
	_, err := wf.mq.pushCommand(
		rrt.CancelWorkflow{RunID: wf.env.WorkflowInfo().WorkflowExecution.RunID},
		nil,
	)

	if err != nil {
		panic(err)
	}
}

// schedule the signal processing
func (wf *workflowProcess) handleSignal(name string, input *commonpb.Payloads) {
	_, err := wf.mq.pushCommand(
		rrt.InvokeSignal{
			RunID: wf.env.WorkflowInfo().WorkflowExecution.RunID,
			Name:  name,
		},
		input,
	)

	if err != nil {
		panic(err)
	}
}

// Handle query in blocking mode.
func (wf *workflowProcess) handleQuery(queryType string, queryArgs *commonpb.Payloads) (*commonpb.Payloads, error) {
	result, err := wf.runCommand(
		rrt.InvokeQuery{
			RunID: wf.runID,
			Name:  queryType,
		},
		queryArgs,
	)

	if err != nil {
		return nil, err
	}

	if result.Failure != nil {
		return nil, bindings.ConvertFailureToError(result.Failure, wf.env.GetDataConverter())
	}

	return result.Payloads, nil
}

// process incoming command
func (wf *workflowProcess) handleMessage(msg rrt.Message) error {
	const op = errors.Op("handleMessage")
	var err error

	var (
		id       = msg.ID
		cmd      = msg.Command
		payloads = msg.Payloads
	)

	switch cmd := cmd.(type) {
	case *rrt.ExecuteActivity:
		params := cmd.ActivityParams(wf.env, payloads)
		activityID := wf.env.ExecuteActivity(params, wf.createCallback(id))

		wf.canceller.register(id, func() error {
			wf.env.RequestCancelActivity(activityID)
			return nil
		})

	case *rrt.ExecuteChildWorkflow:
		params := cmd.WorkflowParams(wf.env, payloads)

		// always use deterministic id
		if params.WorkflowID == "" {
			nextID := atomic.AddUint64(&wf.seqID, 1)
			params.WorkflowID = wf.env.WorkflowInfo().WorkflowExecution.RunID + "_" + strconv.Itoa(int(nextID))
		}

		wf.env.ExecuteChildWorkflow(params, wf.createCallback(id), func(r bindings.WorkflowExecution, e error) {
			wf.ids.push(id, r, e)
		})

		wf.canceller.register(id, func() error {
			wf.env.RequestCancelChildWorkflow(params.Namespace, params.WorkflowID)
			return nil
		})

	case *rrt.GetChildWorkflowExecution:
		wf.ids.listen(cmd.ID, func(w bindings.WorkflowExecution, err error) {
			cl := wf.createCallback(id)

			// TODO rewrite
			if err != nil {
				panic(err)
			}

			p, err := wf.env.GetDataConverter().ToPayloads(w)
			if err != nil {
				panic(err)
			}

			cl(p, err)
		})

	case *rrt.NewTimer:
		timerID := wf.env.NewTimer(cmd.ToDuration(), wf.createCallback(id))
		wf.canceller.register(id, func() error {
			if timerID != nil {
				wf.env.RequestCancelTimer(*timerID)
			}
			return nil
		})

	case *rrt.GetVersion:
		version := wf.env.GetVersion(
			cmd.ChangeID,
			workflow.Version(cmd.MinSupported),
			workflow.Version(cmd.MaxSupported),
		)

		result, err := wf.env.GetDataConverter().ToPayloads(version)
		if err != nil {
			return errors.E(op, err)
		}

		wf.mq.pushResponse(id, result)
		err = wf.flushQueue()
		if err != nil {
			panic(err)
		}

	case *rrt.SideEffect:
		wf.env.SideEffect(
			func() (*commonpb.Payloads, error) {
				return payloads, nil
			},
			wf.createContinuableCallback(id),
		)

	case *rrt.CompleteWorkflow:
		result, _ := wf.env.GetDataConverter().ToPayloads("completed")
		wf.mq.pushResponse(id, result)

		if msg.Failure == nil {
			wf.env.Complete(payloads, nil)
		} else {
			wf.env.Complete(nil, bindings.ConvertFailureToError(msg.Failure, wf.env.GetDataConverter()))
		}

	case *rrt.ContinueAsNew:
		result, _ := wf.env.GetDataConverter().ToPayloads("completed")
		wf.mq.pushResponse(id, result)

		wf.env.Complete(nil, &workflow.ContinueAsNewError{
			WorkflowType:             &bindings.WorkflowType{Name: cmd.Name},
			Input:                    payloads,
			Header:                   wf.header,
			TaskQueueName:            cmd.Options.TaskQueueName,
			WorkflowExecutionTimeout: cmd.Options.WorkflowExecutionTimeout,
			WorkflowRunTimeout:       cmd.Options.WorkflowRunTimeout,
			WorkflowTaskTimeout:      cmd.Options.WorkflowTaskTimeout,
		})

	case *rrt.SignalExternalWorkflow:
		wf.env.SignalExternalWorkflow(
			cmd.Namespace,
			cmd.WorkflowID,
			cmd.RunID,
			cmd.Signal,
			payloads,
			nil,
			cmd.ChildWorkflowOnly,
			wf.createCallback(id),
		)

	case *rrt.CancelExternalWorkflow:
		wf.env.RequestCancelExternalWorkflow(cmd.Namespace, cmd.WorkflowID, cmd.RunID, wf.createCallback(id))

	case *rrt.Cancel:
		err = wf.canceller.cancel(cmd.CommandIDs...)
		if err != nil {
			return errors.E(op, err)
		}

		result, _ := wf.env.GetDataConverter().ToPayloads("completed")
		wf.mq.pushResponse(id, result)

		err = wf.flushQueue()
		if err != nil {
			panic(err)
		}

	case *rrt.Panic:
		panic(errors.E(cmd.Message))

	default:
		panic("undefined command")
	}

	return nil
}

func (wf *workflowProcess) createCallback(id uint64) bindings.ResultHandler {
	callback := func(result *commonpb.Payloads, err error) error {
		wf.canceller.discard(id)

		if err != nil {
			wf.mq.pushError(id, bindings.ConvertErrorToFailure(err, wf.env.GetDataConverter()))
			return nil
		}

		// fetch original payload
		wf.mq.pushResponse(id, result)
		return nil
	}

	return func(result *commonpb.Payloads, err error) {
		// timer cancel callback can happen inside the loop
		if wf.inLoop {
			err := callback(result, err)
			if err != nil {
				panic(err)
			}

			return
		}

		wf.callbacks = append(wf.callbacks, func() error {
			return callback(result, err)
		})
	}
}

// callback to be called inside the queue processing, adds new messages at the end of the queue
func (wf *workflowProcess) createContinuableCallback(id uint64) bindings.ResultHandler {
	callback := func(result *commonpb.Payloads, err error) {
		wf.canceller.discard(id)

		if err != nil {
			wf.mq.pushError(id, bindings.ConvertErrorToFailure(err, wf.env.GetDataConverter()))
			return
		}

		wf.mq.pushResponse(id, result)
		err = wf.flushQueue()
		if err != nil {
			panic(err)
		}
	}

	return func(result *commonpb.Payloads, err error) {
		callback(result, err)
	}
}

// Exchange messages between host and worker processes and add new commands to the queue.
func (wf *workflowProcess) flushQueue() error {
	const op = errors.Op("flush queue")
	messages, err := wf.codec.Execute(wf.pool, wf.getContext(), wf.mq.queue...)
	wf.mq.flush()

	if err != nil {
		return errors.E(op, err)
	}

	wf.pipeline = append(wf.pipeline, messages...)

	return nil
}

// Exchange messages between host and worker processes without adding new commands to the queue.
func (wf *workflowProcess) discardQueue() ([]rrt.Message, error) {
	const op = errors.Op("discard queue")
	messages, err := wf.codec.Execute(wf.pool, wf.getContext(), wf.mq.queue...)
	wf.mq.flush()

	if err != nil {
		return nil, errors.E(op, err)
	}

	return messages, nil
}

// Run single command and return single result.
func (wf *workflowProcess) runCommand(cmd interface{}, payloads *commonpb.Payloads) (rrt.Message, error) {
	const op = errors.Op("run command")
	_, msg, err := wf.mq.allocateMessage(cmd, payloads)
	if err != nil {
		return rrt.Message{}, err
	}

	result, err := wf.codec.Execute(wf.pool, wf.getContext(), msg)
	if err != nil {
		return rrt.Message{}, errors.E(op, err)
	}

	if len(result) != 1 {
		return rrt.Message{}, errors.E("unexpected worker response")
	}

	return result[0], nil
}
