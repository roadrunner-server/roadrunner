package amqp

import (
	"fmt"
	"github.com/spiral/jobs/v2"
	"github.com/streadway/amqp"
)

// pack job metadata into headers
func pack(id string, attempt int, j *jobs.Job) amqp.Table {
	return amqp.Table{
		"rr-id":          id,
		"rr-job":         j.Job,
		"rr-attempt":     int64(attempt),
		"rr-maxAttempts": int64(j.Options.Attempts),
		"rr-timeout":     int64(j.Options.Timeout),
		"rr-delay":       int64(j.Options.Delay),
		"rr-retryDelay":  int64(j.Options.RetryDelay),
	}
}

// unpack restores jobs.Options
func unpack(d amqp.Delivery) (id string, attempt int, j *jobs.Job, err error) {
	j = &jobs.Job{Payload: string(d.Body), Options: &jobs.Options{}}

	if _, ok := d.Headers["rr-id"].(string); !ok {
		return "", 0, nil, fmt.Errorf("missing header `%s`", "rr-id")
	}

	if _, ok := d.Headers["rr-attempt"].(int64); !ok {
		return "", 0, nil, fmt.Errorf("missing header `%s`", "rr-attempt")
	}

	if _, ok := d.Headers["rr-job"].(string); !ok {
		return "", 0, nil, fmt.Errorf("missing header `%s`", "rr-job")
	}
	j.Job = d.Headers["rr-job"].(string)

	if _, ok := d.Headers["rr-maxAttempts"].(int64); ok {
		j.Options.Attempts = int(d.Headers["rr-maxAttempts"].(int64))
	}

	if _, ok := d.Headers["rr-timeout"].(int64); ok {
		j.Options.Timeout = int(d.Headers["rr-timeout"].(int64))
	}

	if _, ok := d.Headers["rr-delay"].(int64); ok {
		j.Options.Delay = int(d.Headers["rr-delay"].(int64))
	}

	if _, ok := d.Headers["rr-retryDelay"].(int64); ok {
		j.Options.RetryDelay = int(d.Headers["rr-retryDelay"].(int64))
	}

	return d.Headers["rr-id"].(string), int(d.Headers["rr-attempt"].(int64)), j, nil
}
