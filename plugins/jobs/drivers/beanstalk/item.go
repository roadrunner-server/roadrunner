package beanstalk

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/beanstalkd/go-beanstalk"
	json "github.com/json-iterator/go"
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

	// Reserve defines for how broker should wait until treating job are failed.
	// - <ttr> -- time to run -- is an integer number of seconds to allow a worker
	// to run this job. This time is counted from the moment a worker reserves
	// this job. If the worker does not delete, release, or bury the job within
	// <ttr> seconds, the job will time out and the server will release the job.
	// The minimum ttr is 1. If the client sends 0, the server will silently
	// increase the ttr to 1. Maximum ttr is 2**32-1.
	Timeout int64 `json:"timeout,omitempty"`

	// Private ================
	id        uint64
	conn      *beanstalk.Conn
	requeueCh chan *Item
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
	return i.Options.conn.Delete(i.Options.id)
}

func (i *Item) Nack() error {
	return i.Options.conn.Delete(i.Options.id)
}

func (i *Item) Requeue(delay int64) error {
	// overwrite the delay
	i.Options.Delay = delay
	select {
	case i.Options.requeueCh <- i:
		return nil
	default:
		return errors.E("can't push to the requeue channel, channel either closed or full", "current size", len(i.Options.requeueCh))
	}
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

func (i *Item) pack(b *bytes.Buffer) error {
	err := gob.NewEncoder(b).Encode(i)
	if err != nil {
		return err
	}

	return nil
}

func (j *JobConsumer) unpack(id uint64, data []byte, out *Item) error {
	err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(out)
	if err != nil {
		return err
	}
	out.Options.conn = j.pool.conn
	out.Options.id = id
	out.Options.requeueCh = j.requeueCh

	return nil
}
