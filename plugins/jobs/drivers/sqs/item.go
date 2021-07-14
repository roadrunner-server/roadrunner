package sqs

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/jobs/job"
	"github.com/spiral/roadrunner/v2/utils"
)

const (
	StringType              string = "String"
	NumberType              string = "Number"
	ApproximateReceiveCount string = "ApproximateReceiveCount"
)

var attributes = []string{
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
}

// CanRetry must return true if broker is allowed to re-run the job.
func (o *Options) CanRetry(attempt int64) bool {
	// Attempts 1 and 0 has identical effect
	return o.MaxAttempts > (attempt + 1)
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

func (j *Item) Priority() int64 {
	return j.Options.Priority
}

// Body packs job payload into binary payload.
func (j *Item) Body() []byte {
	return utils.AsBytes(j.Payload)
}

// Context packs job context (job, id) into binary payload.
// Not used in the sqs, MessageAttributes used instead
func (j *Item) Context() ([]byte, error) {
	return nil, nil
}

func (j *Item) Ack() error {
	return nil
}

func (j *Item) Nack() error {
	return nil
}

func fromJob(job *job.Job) *Item {
	return &Item{
		Job:     job.Job,
		Ident:   job.Ident,
		Payload: job.Payload,
		Options: &Options{
			Priority:    job.Options.Priority,
			Pipeline:    job.Options.Pipeline,
			Delay:       job.Options.Delay,
			Timeout:     job.Options.Timeout,
			MaxAttempts: job.Options.Attempts,
		},
	}
}

func (j *JobConsumer) pack(item *Item) *sqs.SendMessageInput {
	return &sqs.SendMessageInput{
		MessageBody:  aws.String(item.Payload),
		QueueUrl:     j.outputQ.QueueUrl,
		DelaySeconds: int32(item.Options.Delay),
		MessageAttributes: map[string]types.MessageAttributeValue{
			job.RRJob:         {DataType: aws.String(StringType), BinaryValue: nil, BinaryListValues: nil, StringListValues: nil, StringValue: aws.String(item.Job)},
			job.RRDelay:       {DataType: aws.String(StringType), BinaryValue: nil, BinaryListValues: nil, StringListValues: nil, StringValue: aws.String(strconv.Itoa(int(item.Options.Delay)))},
			job.RRTimeout:     {DataType: aws.String(StringType), BinaryValue: nil, BinaryListValues: nil, StringListValues: nil, StringValue: aws.String(strconv.Itoa(int(item.Options.Timeout)))},
			job.RRPriority:    {DataType: aws.String(NumberType), BinaryValue: nil, BinaryListValues: nil, StringListValues: nil, StringValue: aws.String(strconv.Itoa(int(item.Options.Priority)))},
			job.RRMaxAttempts: {DataType: aws.String(NumberType), BinaryValue: nil, BinaryListValues: nil, StringListValues: nil, StringValue: aws.String(strconv.Itoa(int(item.Options.MaxAttempts)))},
		},
	}
}

func (j *JobConsumer) unpack(msg *types.Message) (*Item, int, error) {
	const op = errors.Op("sqs_unpack")
	// reserved
	if _, ok := msg.Attributes[ApproximateReceiveCount]; !ok {
		return nil, 0, errors.E(op, errors.Str("failed to unpack the ApproximateReceiveCount attribute"))
	}

	for i := 0; i < len(attributes); i++ {
		if _, ok := msg.MessageAttributes[attributes[i]]; !ok {
			return nil, 0, errors.E(op, errors.Errorf("missing queue attribute: %s", attributes[i]))
		}
	}

	attempt, err := strconv.Atoi(*msg.MessageAttributes[job.RRMaxAttempts].StringValue)
	if err != nil {
		return nil, 0, errors.E(op, err)
	}

	delay, err := strconv.Atoi(*msg.MessageAttributes[job.RRDelay].StringValue)
	if err != nil {
		return nil, 0, errors.E(op, err)
	}

	to, err := strconv.Atoi(*msg.MessageAttributes[job.RRTimeout].StringValue)
	if err != nil {
		return nil, 0, errors.E(op, err)
	}

	priority, err := strconv.Atoi(*msg.MessageAttributes[job.RRPriority].StringValue)
	if err != nil {
		return nil, 0, errors.E(op, err)
	}

	recCount, err := strconv.Atoi(msg.Attributes[ApproximateReceiveCount])
	if err != nil {
		return nil, 0, errors.E(op, err)
	}

	item := &Item{
		Job:     *msg.MessageAttributes[job.RRJob].StringValue,
		Payload: *msg.Body,
		Options: &Options{
			Delay:       int64(delay),
			Timeout:     int64(to),
			Priority:    int64(priority),
			MaxAttempts: int64(attempt),
		},
	}

	return item, recCount, nil
}
