package protocol

import (
	"github.com/spiral/errors"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"
)

// WorkerInfo outlines information about every available worker and it's TaskQueues.

// WorkerInfo lists available task queues, workflows and activities.
type WorkerInfo struct {
	// TaskQueue assigned to the worker.
	TaskQueue string `json:"taskQueue"`

	// Options describe worker options.
	Options worker.Options `json:"options,omitempty"`

	// Workflows provided by the worker.
	Workflows []WorkflowInfo

	// Activities provided by the worker.
	Activities []ActivityInfo
}

// WorkflowInfo describes single worker workflow.
type WorkflowInfo struct {
	// Name of the workflow.
	Name string `json:"name"`

	// Queries pre-defined for the workflow type.
	Queries []string `json:"queries"`

	// Signals pre-defined for the workflow type.
	Signals []string `json:"signals"`
}

// ActivityInfo describes single worker activity.
type ActivityInfo struct {
	// Name describes public activity name.
	Name string `json:"name"`
}

// FetchWorkerInfo fetches information about all underlying workers (can be multiplexed inside single process).
func FetchWorkerInfo(c Codec, e Endpoint, dc converter.DataConverter) ([]WorkerInfo, error) {
	const op = errors.Op("fetch_worker_info")

	result, err := c.Execute(e, Context{}, Message{ID: 0, Command: GetWorkerInfo{}})
	if err != nil {
		return nil, errors.E(op, err)
	}

	if len(result) != 1 {
		return nil, errors.E(op, errors.Str("unable to read worker info"))
	}

	if result[0].ID != 0 {
		return nil, errors.E(op, errors.Str("FetchWorkerInfo confirmation missing"))
	}

	var info []WorkerInfo
	for i := range result[0].Payloads.Payloads {
		wi := WorkerInfo{}
		if err := dc.FromPayload(result[0].Payloads.Payloads[i], &wi); err != nil {
			return nil, errors.E(op, err)
		}

		info = append(info, wi)
	}

	return info, nil
}
