package sqs

import (
	"context"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	json "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/jobs/job"
	"github.com/spiral/roadrunner/v2/utils"
)

const (
	StringType              string = "String"
	NumberType              string = "Number"
	ApproximateReceiveCount string = "ApproximateReceiveCount"
)

var itemAttributes = []string{
	job.RRJob,
	job.RRDelay,
	job.RRTimeout,
	job.RRPriority,
	job.RRMaxAttempts,
}

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

	// Maximum number of attempts to receive and process the message
	MaxAttempts int64 `json:"max_attempts,omitempty"`

	// Private ================
	approxReceiveCount int64
	queue              *string
	receiptHandler     *string
	client             *sqs.Client
}

// CanRetry must return true if broker is allowed to re-run the job.
func (o *Options) CanRetry() bool {
	// Attempts 1 and 0 has identical effect
	if o.MaxAttempts == 0 || o.MaxAttempts == 1 {
		return false
	}
	return o.MaxAttempts > (o.approxReceiveCount + 1)
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
// Not used in the sqs, MessageAttributes used instead
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
	_, err := i.Options.client.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
		QueueUrl:      i.Options.queue,
		ReceiptHandle: i.Options.receiptHandler,
	})

	if err != nil {
		return err
	}

	return nil
}

func (i *Item) Nack() error {
	if i.Options.CanRetry() {
		return nil
	}

	_, err := i.Options.client.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
		QueueUrl:      i.Options.queue,
		ReceiptHandle: i.Options.receiptHandler,
	})

	if err != nil {
		return err
	}

	return nil
}

func fromJob(job *job.Job) *Item {
	return &Item{
		Job:     job.Job,
		Ident:   job.Ident,
		Payload: job.Payload,
		Headers: job.Headers,
		Options: &Options{
			Priority:    job.Options.Priority,
			Pipeline:    job.Options.Pipeline,
			Delay:       job.Options.Delay,
			Timeout:     job.Options.Timeout,
			MaxAttempts: job.Options.Attempts,
		},
	}
}

func (i *Item) pack(queue *string) *sqs.SendMessageInput {
	return &sqs.SendMessageInput{
		MessageBody:  aws.String(i.Payload),
		QueueUrl:     queue,
		DelaySeconds: int32(i.Options.Delay),
		MessageAttributes: map[string]types.MessageAttributeValue{
			job.RRJob:         {DataType: aws.String(StringType), BinaryValue: nil, BinaryListValues: nil, StringListValues: nil, StringValue: aws.String(i.Job)},
			job.RRDelay:       {DataType: aws.String(StringType), BinaryValue: nil, BinaryListValues: nil, StringListValues: nil, StringValue: aws.String(strconv.Itoa(int(i.Options.Delay)))},
			job.RRTimeout:     {DataType: aws.String(StringType), BinaryValue: nil, BinaryListValues: nil, StringListValues: nil, StringValue: aws.String(strconv.Itoa(int(i.Options.Timeout)))},
			job.RRPriority:    {DataType: aws.String(NumberType), BinaryValue: nil, BinaryListValues: nil, StringListValues: nil, StringValue: aws.String(strconv.Itoa(int(i.Options.Priority)))},
			job.RRMaxAttempts: {DataType: aws.String(NumberType), BinaryValue: nil, BinaryListValues: nil, StringListValues: nil, StringValue: aws.String(strconv.Itoa(int(i.Options.MaxAttempts)))},
		},
	}
}

func unpack(msg *types.Message, queue *string, client *sqs.Client) (*Item, error) {
	const op = errors.Op("sqs_unpack")
	// reserved
	if _, ok := msg.Attributes[ApproximateReceiveCount]; !ok {
		return nil, errors.E(op, errors.Str("failed to unpack the ApproximateReceiveCount attribute"))
	}

	for i := 0; i < len(itemAttributes); i++ {
		if _, ok := msg.MessageAttributes[itemAttributes[i]]; !ok {
			return nil, errors.E(op, errors.Errorf("missing queue attribute: %s", itemAttributes[i]))
		}
	}

	attempt, err := strconv.Atoi(*msg.MessageAttributes[job.RRMaxAttempts].StringValue)
	if err != nil {
		return nil, errors.E(op, err)
	}

	delay, err := strconv.Atoi(*msg.MessageAttributes[job.RRDelay].StringValue)
	if err != nil {
		return nil, errors.E(op, err)
	}

	to, err := strconv.Atoi(*msg.MessageAttributes[job.RRTimeout].StringValue)
	if err != nil {
		return nil, errors.E(op, err)
	}

	priority, err := strconv.Atoi(*msg.MessageAttributes[job.RRPriority].StringValue)
	if err != nil {
		return nil, errors.E(op, err)
	}

	recCount, err := strconv.Atoi(msg.Attributes[ApproximateReceiveCount])
	if err != nil {
		return nil, errors.E(op, err)
	}

	item := &Item{
		Job:     *msg.MessageAttributes[job.RRJob].StringValue,
		Payload: *msg.Body,
		Options: &Options{
			Delay:       int64(delay),
			Timeout:     int64(to),
			Priority:    int64(priority),
			MaxAttempts: int64(attempt),

			// private
			approxReceiveCount: int64(recCount),
			client:             client,
			queue:              queue,
			receiptHandler:     msg.ReceiptHandle,
		},
	}

	return item, nil
}
