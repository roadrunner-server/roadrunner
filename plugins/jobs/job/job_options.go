package job

import "time"

// Options carry information about how to handle given job.
type Options struct {
	// Priority is job priority, default - 10
	// pointer to distinguish 0 as a priority and nil as priority not set
	Priority int64 `json:"priority"`

	// Pipeline manually specified pipeline.
	Pipeline string `json:"pipeline,omitempty"`

	// Delay defines time duration to delay execution for. Defaults to none.
	Delay int64 `json:"delay,omitempty"`
}

// Merge merges job options.
func (o *Options) Merge(from *Options) {
	if o.Pipeline == "" {
		o.Pipeline = from.Pipeline
	}

	if o.Delay == 0 {
		o.Delay = from.Delay
	}
}

// DelayDuration returns delay duration in a form of time.Duration.
func (o *Options) DelayDuration() time.Duration {
	return time.Second * time.Duration(o.Delay)
}
