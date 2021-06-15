package jobs

import json "github.com/json-iterator/go"

// Handler handles job execution.
type Handler func(id string, j *Job) error

// ErrorHandler handles job execution errors.
type ErrorHandler func(id string, j *Job, err error)

// Job carries information about single job.
type Job struct {
	// Job contains name of job broker (usually PHP class).
	Job string `json:"job"`

	// Payload is string data (usually JSON) passed to Job broker.
	Payload string `json:"payload"`

	// Options contains set of PipelineOptions specific to job execution. Can be empty.
	Options *Options `json:"options,omitempty"`
}

// Body packs job payload into binary payload.
func (j *Job) Body() []byte {
	return []byte(j.Payload)
}

// Context packs job context (job, id) into binary payload.
func (j *Job) Context(id string) []byte {
	ctx, _ := json.Marshal(
		struct {
			ID  string `json:"id"`
			Job string `json:"job"`
		}{ID: id, Job: j.Job},
	)

	return ctx
}
