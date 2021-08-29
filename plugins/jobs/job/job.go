package job

import (
	"time"
)

// constant keys to pack/unpack messages from different drivers
const (
	RRID       string = "rr_id"
	RRJob      string = "rr_job"
	RRHeaders  string = "rr_headers"
	RRPipeline string = "rr_pipeline"
	RRDelay    string = "rr_delay"
	RRPriority string = "rr_priority"
)

// Job carries information about single job.
type Job struct {
	// Job contains name of job broker (usually PHP class).
	Job string `json:"job"`

	// Ident is unique identifier of the job, should be provided from outside
	Ident string `json:"id"`

	// Payload is string data (usually JSON) passed to Job broker.
	Payload string `json:"payload"`

	// Headers with key-value pairs
	Headers map[string][]string `json:"headers"`

	// Options contains set of PipelineOptions specific to job execution. Can be empty.
	Options *Options `json:"options,omitempty"`
}

// Options carry information about how to handle given job.
type Options struct {
	// Priority is job priority, default - 10
	// pointer to distinguish 0 as a priority and nil as priority not set
	Priority int64 `json:"priority"`

	// Pipeline manually specified pipeline.
	Pipeline string `json:"pipeline,omitempty"`

	// Delay defines time duration to delay execution for. Defaults to none.
	Delay int64 `json:"delay,omitempty"`
}

// Merge merges job options.
func (o *Options) Merge(from *Options) {
	if o.Pipeline == "" {
		o.Pipeline = from.Pipeline
	}

	if o.Delay == 0 {
		o.Delay = from.Delay
	}
}

// DelayDuration returns delay duration in a form of time.Duration.
func (o *Options) DelayDuration() time.Duration {
	return time.Second * time.Duration(o.Delay)
}
