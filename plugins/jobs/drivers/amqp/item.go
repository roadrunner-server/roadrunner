package amqp

import (
	"time"

	json "github.com/json-iterator/go"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/jobs/job"
	"github.com/spiral/roadrunner/v2/utils"
)

type Item struct {
	// Job contains pluginName of job broker (usually PHP class).
	Job string `json:"job"`

	// Ident is unique identifier of the job, should be provided from outside
	Ident string `json:"id"`

	// Payload is string data (usually JSON) passed to Job broker.
	Payload string `json:"payload"`

	// Headers with key-values pairs
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

	// Reserve defines for how broker should wait until treating job are failed. Defaults to 30 min.
	Timeout int64 `json:"timeout,omitempty"`

	// private
	// Ack delegates an acknowledgement through the Acknowledger interface that the client or server has finished work on a delivery
	ack func(multiply bool) error

	// Nack negatively acknowledge the delivery of message(s) identified by the delivery tag from either the client or server.
	// When multiple is true, nack messages up to and including delivered messages up until the delivery tag delivered on the same channel.
	// When requeue is true, request the server to deliver this message to a different consumer. If it is not possible or requeue is false, the message will be dropped or delivered to a server configured dead-letter queue.
	// This method must not be used to select or requeue messages the client wishes not to handle, rather it is to inform the server that the client is incapable of handling this message at this time
	nack func(multiply bool, requeue bool) error

	requeueCh chan *Item

	multipleAsk bool
	requeue     bool
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

func (i *Item) ID() string {
	return i.Ident
}

func (i *Item) Priority() int64 {
	return i.Options.Priority
}

// Body packs job payload into binary payload.
func (i *Item) Body() []byte {
	return utils.AsBytes(i.Payload)
}

// Context packs job context (job, id) into binary payload.
// Not used in the amqp, amqp.Table used instead
func (i *Item) Context() ([]byte, error) {
	ctx, err := json.Marshal(
		struct {
			ID       string              `json:"id"`
			Job      string              `json:"job"`
			Headers  map[string][]string `json:"headers"`
			Timeout  int64               `json:"timeout"`
			Pipeline string              `json:"pipeline"`
		}{ID: i.Ident, Job: i.Job, Headers: i.Headers, Timeout: i.Options.Timeout, Pipeline: i.Options.Pipeline},
	)

	if err != nil {
		return nil, err
	}

	return ctx, nil
}

func (i *Item) Ack() error {
	return i.Options.ack(i.Options.multipleAsk)
}

func (i *Item) Nack() error {
	return i.Options.nack(false, i.Options.requeue)
}

// Requeue with the provided delay, handled by the Nack
func (i *Item) Requeue(headers map[string][]string, delay int64) error {
	// overwrite the delay
	i.Options.Delay = delay
	i.Headers = headers
	select {
	case i.Options.requeueCh <- i:
		return nil
	default:
		return errors.E("can't push to the requeue channel, channel either closed or full", "current size", len(i.Options.requeueCh))
	}
}

// fromDelivery converts amqp.Delivery into an Item which will be pushed to the PQ
func (j *JobConsumer) fromDelivery(d amqp.Delivery) (*Item, error) {
	const op = errors.Op("from_delivery_convert")
	item, err := j.unpack(d)
	if err != nil {
		return nil, errors.E(op, err)
	}

	i := &Item{
		Job:     item.Job,
		Ident:   item.Ident,
		Payload: item.Payload,
		Headers: item.Headers,
		Options: item.Options,
	}

	item.Options.ack = d.Ack
	item.Options.nack = d.Nack
	// requeue channel
	item.Options.requeueCh = j.requeueCh
	return i, nil
}

func fromJob(job *job.Job) *Item {
	return &Item{
		Job:     job.Job,
		Ident:   job.Ident,
		Payload: job.Payload,
		Headers: job.Headers,
		Options: &Options{
			Priority: job.Options.Priority,
			Pipeline: job.Options.Pipeline,
			Delay:    job.Options.Delay,
			Timeout:  job.Options.Timeout,
		},
	}
}

// pack job metadata into headers
func pack(id string, j *Item) (amqp.Table, error) {
	headers, err := json.Marshal(j.Headers)
	if err != nil {
		return nil, err
	}
	return amqp.Table{
		job.RRID:       id,
		job.RRJob:      j.Job,
		job.RRPipeline: j.Options.Pipeline,
		job.RRHeaders:  headers,
		job.RRTimeout:  j.Options.Timeout,
		job.RRDelay:    j.Options.Delay,
		job.RRPriority: j.Options.Priority,
	}, nil
}

// unpack restores jobs.Options
func (j *JobConsumer) unpack(d amqp.Delivery) (*Item, error) {
	item := &Item{Payload: utils.AsString(d.Body), Options: &Options{
		multipleAsk: j.multipleAck,
		requeue:     j.requeueOnFail,
		requeueCh:   j.requeueCh,
	}}

	if _, ok := d.Headers[job.RRID].(string); !ok {
		return nil, errors.E(errors.Errorf("missing header `%s`", job.RRID))
	}

	item.Ident = d.Headers[job.RRID].(string)

	if _, ok := d.Headers[job.RRJob].(string); !ok {
		return nil, errors.E(errors.Errorf("missing header `%s`", job.RRJob))
	}

	item.Job = d.Headers[job.RRJob].(string)

	if _, ok := d.Headers[job.RRPipeline].(string); ok {
		item.Options.Pipeline = d.Headers[job.RRPipeline].(string)
	}

	if h, ok := d.Headers[job.RRHeaders].([]byte); ok {
		err := json.Unmarshal(h, &item.Headers)
		if err != nil {
			return nil, err
		}
	}

	if _, ok := d.Headers[job.RRTimeout].(int64); ok {
		item.Options.Timeout = d.Headers[job.RRTimeout].(int64)
	}

	if _, ok := d.Headers[job.RRDelay].(int64); ok {
		item.Options.Delay = d.Headers[job.RRDelay].(int64)
	}

	if _, ok := d.Headers[job.RRPriority]; !ok {
		// set pipe's priority
		item.Options.Priority = j.priority
	} else {
		item.Options.Priority = d.Headers[job.RRPriority].(int64)
	}

	return item, nil
}
