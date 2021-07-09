package amqp

import (
	"fmt"

	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
	"github.com/streadway/amqp"
)

const (
	rrID          string = "rr-id"
	rrJob         string = "rr-job"
	rrAttempt     string = "rr-attempt"
	rrMaxAttempts string = "rr-max_attempts"
	rrTimeout     string = "rr-timeout"
	rrDelay       string = "rr-delay"
	rrRetryDelay  string = "rr-retry_delay"
)

// pack job metadata into headers
func pack(id string, attempt uint64, j *structs.Job) amqp.Table {
	return amqp.Table{
		rrID:          id,
		rrJob:         j.Job,
		rrAttempt:     attempt,
		rrMaxAttempts: j.Options.Attempts,
		rrTimeout:     j.Options.Timeout,
		rrDelay:       j.Options.Delay,
		rrRetryDelay:  j.Options.RetryDelay,
	}
}

// unpack restores jobs.Options
func unpack(d amqp.Delivery) (id string, attempt int, j *structs.Job, err error) { //nolint:deadcode,unused
	j = &structs.Job{Payload: string(d.Body), Options: &structs.Options{}}

	if _, ok := d.Headers[rrID].(string); !ok {
		return "", 0, nil, fmt.Errorf("missing header `%s`", rrID)
	}

	if _, ok := d.Headers[rrAttempt].(uint64); !ok {
		return "", 0, nil, fmt.Errorf("missing header `%s`", rrAttempt)
	}

	if _, ok := d.Headers[rrJob].(string); !ok {
		return "", 0, nil, fmt.Errorf("missing header `%s`", rrJob)
	}

	j.Job = d.Headers[rrJob].(string)

	if _, ok := d.Headers[rrMaxAttempts].(uint64); ok {
		j.Options.Attempts = d.Headers[rrMaxAttempts].(uint64)
	}

	if _, ok := d.Headers[rrTimeout].(uint64); ok {
		j.Options.Timeout = d.Headers[rrTimeout].(uint64)
	}

	if _, ok := d.Headers[rrDelay].(uint64); ok {
		j.Options.Delay = d.Headers[rrDelay].(uint64)
	}

	if _, ok := d.Headers[rrRetryDelay].(uint64); ok {
		j.Options.RetryDelay = d.Headers[rrRetryDelay].(uint64)
	}

	return d.Headers[rrID].(string), int(d.Headers[rrAttempt].(uint64)), j, nil
}
