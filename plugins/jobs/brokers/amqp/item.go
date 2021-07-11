package amqp

import (
	"fmt"
	"time"

	json "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
	"github.com/spiral/roadrunner/v2/utils"
	"github.com/streadway/amqp"
)

const (
	rrID         string = "rr_id"
	rrJob        string = "rr_job"
	rrHeaders    string = "rr_headers"
	rrPipeline   string = "rr_pipeline"
	rrTimeout    string = "rr_timeout"
	rrDelay      string = "rr_delay"
	rrRetryDelay string = "rr_retry_delay"
)

func FromDelivery(d amqp.Delivery) (*Item, error) {
	const op = errors.Op("from_delivery_convert")
	id, item, err := unpack(d)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return &Item{
		Job:      item.Job,
		Ident:    id,
		Payload:  item.Payload,
		Headers:  item.Headers,
		Options:  item.Options,
		AckFunc:  d.Ack,
		NackFunc: d.Nack,
	}, nil
}

func FromJob(job *structs.Job) *Item {
	return &Item{
		Job:     job.Job,
		Ident:   job.Ident,
		Payload: job.Payload,
		Options: &Options{
			Priority:   uint32(job.Options.Priority),
			Pipeline:   job.Options.Pipeline,
			Delay:      int32(job.Options.Delay),
			Attempts:   int32(job.Options.Attempts),
			RetryDelay: int32(job.Options.RetryDelay),
			Timeout:    int32(job.Options.Timeout),
		},
	}
}

type Item struct {
	// Job contains pluginName of job broker (usually PHP class).
	Job string `json:"job"`

	// Ident is unique identifier of the job, should be provided from outside
	Ident string

	// Payload is string data (usually JSON) passed to Job broker.
	Payload string `json:"payload"`

	// Headers with key-values pairs
	Headers map[string][]string

	// Options contains set of PipelineOptions specific to job execution. Can be empty.
	Options *Options `json:"options,omitempty"`

	// Ack delegates an acknowledgement through the Acknowledger interface that the client or server has finished work on a delivery
	AckFunc func(multiply bool) error

	// Nack negatively acknowledge the delivery of message(s) identified by the delivery tag from either the client or server.
	// When multiple is true, nack messages up to and including delivered messages up until the delivery tag delivered on the same channel.
	// When requeue is true, request the server to deliver this message to a different consumer. If it is not possible or requeue is false, the message will be dropped or delivered to a server configured dead-letter queue.
	// This method must not be used to select or requeue messages the client wishes not to handle, rather it is to inform the server that the client is incapable of handling this message at this time
	NackFunc func(multiply bool, requeue bool) error
}

// Options carry information about how to handle given job.
type Options struct {
	// Priority is job priority, default - 10
	// pointer to distinguish 0 as a priority and nil as priority not set
	Priority uint32 `json:"priority"`

	// Pipeline manually specified pipeline.
	Pipeline string `json:"pipeline,omitempty"`

	// Delay defines time duration to delay execution for. Defaults to none.
	Delay int32 `json:"delay,omitempty"`

	// Attempts define maximum job retries. Attention, value 1 will only allow job to execute once (without retry).
	// Minimum valuable value is 2.
	Attempts int32 `json:"maxAttempts,omitempty"`

	// RetryDelay defines for how long job should be waiting until next retry. Defaults to none.
	RetryDelay int32 `json:"retryDelay,omitempty"`

	// Reserve defines for how broker should wait until treating job are failed. Defaults to 30 min.
	Timeout int32 `json:"timeout,omitempty"`
}

// CanRetry must return true if broker is allowed to re-run the job.
func (o *Options) CanRetry(attempt int32) bool {
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
	return uint64(j.Options.Priority)
}

// Body packs job payload into binary payload.
func (j *Item) Body() []byte {
	return utils.AsBytes(j.Payload)
}

// Context packs job context (job, id) into binary payload.
// Not used in the amqp, amqp.Table used instead
func (j *Item) Context() ([]byte, error) {
	return nil, nil
}

func (j *Item) Ack() error {
	return j.AckFunc(false)
}

func (j *Item) Nack() error {
	return j.NackFunc(false, false)
}

// pack job metadata into headers
func pack(id string, j *Item) (amqp.Table, error) {
	headers, err := json.Marshal(j.Headers)
	if err != nil {
		return nil, err
	}
	return amqp.Table{
		rrID:         id,
		rrJob:        j.Job,
		rrPipeline:   j.Options.Pipeline,
		rrHeaders:    headers,
		rrTimeout:    j.Options.Timeout,
		rrDelay:      j.Options.Delay,
		rrRetryDelay: j.Options.RetryDelay,
	}, nil
}

// unpack restores jobs.Options
func unpack(d amqp.Delivery) (id string, j *Item, err error) {
	j = &Item{Payload: string(d.Body), Options: &Options{}}

	if _, ok := d.Headers[rrID].(string); !ok {
		return "", nil, fmt.Errorf("missing header `%s`", rrID)
	}

	if _, ok := d.Headers[rrJob].(string); !ok {
		return "", nil, fmt.Errorf("missing header `%s`", rrJob)
	}

	j.Job = d.Headers[rrJob].(string)

	if _, ok := d.Headers[rrPipeline].(string); ok {
		j.Options.Pipeline = d.Headers[rrPipeline].(string)
	}

	if h, ok := d.Headers[rrHeaders].([]byte); ok {
		err := json.Unmarshal(h, &j.Headers)
		if err != nil {
			return "", nil, err
		}
	}

	if _, ok := d.Headers[rrTimeout].(int32); ok {
		j.Options.Timeout = d.Headers[rrTimeout].(int32)
	}

	if _, ok := d.Headers[rrDelay].(int32); ok {
		j.Options.Delay = d.Headers[rrDelay].(int32)
	}

	if _, ok := d.Headers[rrRetryDelay].(int32); ok {
		j.Options.RetryDelay = d.Headers[rrRetryDelay].(int32)
	}

	return d.Headers[rrID].(string), j, nil
}
