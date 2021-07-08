package ephemeral

import (
	"time"

	json "github.com/json-iterator/go"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
	"github.com/spiral/roadrunner/v2/utils"
)

func From(job *structs.Job) *Item {
	return &Item{
		Job:     job.Job,
		Ident:   job.Ident,
		Payload: job.Payload,
		Options: conv(*job.Options),
	}
}

func conv(jo structs.Options) Options {
	return Options(jo)
}

type Item struct {
	// Job contains name of job broker (usually PHP class).
	Job string `json:"job"`

	// Ident is unique identifier of the job, should be provided from outside
	Ident string

	// Payload is string data (usually JSON) passed to Job broker.
	Payload string `json:"payload"`

	// Headers with key-values pairs
	Headers map[string][]string

	// Options contains set of PipelineOptions specific to job execution. Can be empty.
	Options Options `json:"options,omitempty"`
}

// Options carry information about how to handle given job.
type Options struct {
	// Priority is job priority, default - 10
	// pointer to distinguish 0 as a priority and nil as priority not set
	Priority uint64 `json:"priority"`

	// Pipeline manually specified pipeline.
	Pipeline string `json:"pipeline,omitempty"`

	// Delay defines time duration to delay execution for. Defaults to none.
	Delay uint64 `json:"delay,omitempty"`

	// Attempts define maximum job retries. Attention, value 1 will only allow job to execute once (without retry).
	// Minimum valuable value is 2.
	Attempts uint64 `json:"maxAttempts,omitempty"`

	// RetryDelay defines for how long job should be waiting until next retry. Defaults to none.
	RetryDelay uint64 `json:"retryDelay,omitempty"`

	// Reserve defines for how broker should wait until treating job are failed. Defaults to 30 min.
	Timeout uint64 `json:"timeout,omitempty"`
}

// CanRetry must return true if broker is allowed to re-run the job.
func (o *Options) CanRetry(attempt uint64) bool {
	// Attempts 1 and 0 has identical effect
	return o.Attempts > (attempt + 1)
}

// RetryDuration returns retry delay duration in a form of time.Duration.
func (o *Options) RetryDuration() time.Duration {
	return time.Second * time.Duration(o.RetryDelay)
}

// DelayDuration returns delay duration in a form of time.Duration.
func (o *Options) DelayDuration() time.Duration {
	return time.Second * time.Duration(o.Delay)
}

// TimeoutDuration returns timeout duration in a form of time.Duration.
func (o *Options) TimeoutDuration() time.Duration {
	if o.Timeout == 0 {
		return 30 * time.Minute
	}

	return time.Second * time.Duration(o.Timeout)
}

func (j *Item) ID() string {
	return j.Ident
}

func (j *Item) Priority() uint64 {
	return j.Options.Priority
}

// Body packs job payload into binary payload.
func (j *Item) Body() []byte {
	return utils.AsBytes(j.Payload)
}

// Context packs job context (job, id) into binary payload.
func (j *Item) Context() ([]byte, error) {
	ctx, err := json.Marshal(
		struct {
			ID       string              `json:"id"`
			Job      string              `json:"job"`
			Headers  map[string][]string `json:"headers"`
			Timeout  uint64              `json:"timeout"`
			Pipeline string              `json:"pipeline"`
		}{ID: j.Ident, Job: j.Job, Headers: j.Headers, Timeout: j.Options.Timeout, Pipeline: j.Options.Pipeline},
	)

	if err != nil {
		return nil, err
	}

	return ctx, nil
}

func (j *Item) Ack() error {
	// noop for the in-memory
	return nil
}

func (j *Item) Nack() error {
	// noop for the in-memory
	return nil
}
