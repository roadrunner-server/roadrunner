package job

// constant keys to pack/unpack messages from different drivers
const (
	RRID          string = "rr_id"
	RRJob         string = "rr_job"
	RRHeaders     string = "rr_headers"
	RRPipeline    string = "rr_pipeline"
	RRTimeout     string = "rr_timeout"
	RRDelay       string = "rr_delay"
	RRPriority    string = "rr_priority"
	RRMaxAttempts string = "rr_max_attempts"
)

// Job carries information about single job.
type Job struct {
	// Job contains name of job broker (usually PHP class).
	Job string `json:"job"`

	// Ident is unique identifier of the job, should be provided from outside
	Ident string `json:"id"`

	// Payload is string data (usually JSON) passed to Job broker.
	Payload string `json:"payload"`

	// Headers with key-value pairs
	Headers map[string][]string `json:"headers"`

	// Options contains set of PipelineOptions specific to job execution. Can be empty.
	Options *Options `json:"options,omitempty"`
}

func (j Job) ID() string {
	panic("implement me")
}

func (j Job) Priority() int64 {
	panic("implement me")
}

func (j Job) Body() []byte {
	panic("implement me")
}

func (j Job) Context() ([]byte, error) {
	panic("implement me")
}

func (j Job) Ack() error {
	panic("implement me")
}

func (j Job) Nack() error {
	panic("implement me")
}

func (j Job) Requeue(delay uint32) error {
	panic("implement me")
}
