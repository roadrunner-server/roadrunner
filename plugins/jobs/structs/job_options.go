package structs

import "time"

// Options carry information about how to handle given job.
type Options struct {
	// Priority is job priority, default - 10
	// pointer to distinguish 0 as a priority and nil as priority not set
	Priority *uint64 `json:"priority"`

	// ID - generated ID for the job
	ID string `json:"id"`

	// Pipeline manually specified pipeline.
	Pipeline string `json:"pipeline,omitempty"`

	// Delay defines time duration to delay execution for. Defaults to none.
	Delay int `json:"delay,omitempty"`

	// Attempts define maximum job retries. Attention, value 1 will only allow job to execute once (without retry).
	// Minimum valuable value is 2.
	Attempts int `json:"maxAttempts,omitempty"`

	// RetryDelay defines for how long job should be waiting until next retry. Defaults to none.
	RetryDelay int `json:"retryDelay,omitempty"`

	// Reserve defines for how broker should wait until treating job are failed. Defaults to 30 min.
	Timeout int `json:"timeout,omitempty"`
}

// Merge merges job options.
func (o *Options) Merge(from *Options) {
	if o.Pipeline == "" {
		o.Pipeline = from.Pipeline
	}

	if o.Attempts == 0 {
		o.Attempts = from.Attempts
	}

	if o.Timeout == 0 {
		o.Timeout = from.Timeout
	}

	if o.RetryDelay == 0 {
		o.RetryDelay = from.RetryDelay
	}

	if o.Delay == 0 {
		o.Delay = from.Delay
	}
}

// CanRetry must return true if broker is allowed to re-run the job.
func (o *Options) CanRetry(attempt int) bool {
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
