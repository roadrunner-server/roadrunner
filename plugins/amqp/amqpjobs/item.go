package amqpjobs

import (
	"context"
	"fmt"
	"sync/atomic"
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

	// private
	// ack delegates an acknowledgement through the Acknowledger interface that the client or server has finished work on a delivery
	ack func(multiply bool) error

	// nack negatively acknowledge the delivery of message(s) identified by the delivery tag from either the client or server.
	// When multiple is true, nack messages up to and including delivered messages up until the delivery tag delivered on the same channel.
	// When requeue is true, request the server to deliver this message to a different consumer. If it is not possible or requeue is false, the message will be dropped or delivered to a server configured dead-letter queue.
	// This method must not be used to select or requeue messages the client wishes not to handle, rather it is to inform the server that the client is incapable of handling this message at this time
	nack func(multiply bool, requeue bool) error

	// requeueFn used as a pointer to the push function
	requeueFn func(context.Context, *Item) error
	// delayed jobs TODO(rustatian): figure out how to get stats from the DLX
	delayed     *int64
	multipleAsk bool
	requeue     bool
}

// DelayDuration returns delay duration in a form of time.Duration.
func (o *Options) DelayDuration() time.Duration {
	return time.Second * time.Duration(o.Delay)
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
			Pipeline string              `json:"pipeline"`
		}{ID: i.Ident, Job: i.Job, Headers: i.Headers, Pipeline: i.Options.Pipeline},
	)

	if err != nil {
		return nil, err
	}

	return ctx, nil
}

func (i *Item) Ack() error {
	if i.Options.Delay > 0 {
		atomic.AddInt64(i.Options.delayed, ^int64(0))
	}
	return i.Options.ack(i.Options.multipleAsk)
}

func (i *Item) Nack() error {
	if i.Options.Delay > 0 {
		atomic.AddInt64(i.Options.delayed, ^int64(0))
	}
	return i.Options.nack(false, i.Options.requeue)
}

// Requeue with the provided delay, handled by the Nack
func (i *Item) Requeue(headers map[string][]string, delay int64) error {
	if i.Options.Delay > 0 {
		atomic.AddInt64(i.Options.delayed, ^int64(0))
	}
	// overwrite the delay
	i.Options.Delay = delay
	i.Headers = headers

	err := i.Options.requeueFn(context.Background(), i)
	if err != nil {
		errNack := i.Options.nack(false, true)
		if errNack != nil {
			return fmt.Errorf("requeue error: %v\nack error: %v", err, errNack)
		}

		return err
	}

	// ack the job
	err = i.Options.ack(false)
	if err != nil {
		return err
	}

	return nil
}

// fromDelivery converts amqp.Delivery into an Item which will be pushed to the PQ
func (c *consumer) fromDelivery(d amqp.Delivery) (*Item, error) {
	const op = errors.Op("from_delivery_convert")
	item, err := c.unpack(d)
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
	item.Options.delayed = c.delayed

	// requeue func
	item.Options.requeueFn = c.handleItem
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
		job.RRDelay:    j.Options.Delay,
		job.RRPriority: j.Options.Priority,
	}, nil
}

// unpack restores jobs.Options
func (c *consumer) unpack(d amqp.Delivery) (*Item, error) {
	item := &Item{Payload: utils.AsString(d.Body), Options: &Options{
		multipleAsk: c.multipleAck,
		requeue:     c.requeueOnFail,
		requeueFn:   c.handleItem,
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

	if t, ok := d.Headers[job.RRDelay]; ok {
		switch t.(type) {
		case int, int16, int32, int64:
			item.Options.Delay = t.(int64)
		default:
			c.log.Warn("unknown delay type", "want:", "int, int16, int32, int64", "actual", t)
		}
	}

	if t, ok := d.Headers[job.RRPriority]; !ok {
		// set pipe's priority
		item.Options.Priority = c.priority
	} else {
		switch t.(type) {
		case int, int16, int32, int64:
			item.Options.Priority = t.(int64)
		default:
			c.log.Warn("unknown priority type", "want:", "int, int16, int32, int64", "actual", t)
		}
	}

	return item, nil
}
