package structs

import (
	json "github.com/json-iterator/go"
	"github.com/spiral/roadrunner/v2/utils"
)

// Job carries information about single job.
type Job struct {
	// Job contains name of job broker (usually PHP class).
	Job string `json:"job"`

	// Payload is string data (usually JSON) passed to Job broker.
	Payload string `json:"payload"`

	// Options contains set of PipelineOptions specific to job execution. Can be empty.
	Options *Options `json:"options,omitempty"`
}

func (j *Job) ID() string {
	return j.Options.ID
}

func (j *Job) Priority() uint64 {
	return *j.Options.Priority
}

// Body packs job payload into binary payload.
func (j *Job) Body() []byte {
	return utils.AsBytes(j.Payload)
}

// Context packs job context (job, id) into binary payload.
func (j *Job) Context() []byte {
	ctx, _ := json.Marshal(
		struct {
			ID  string `json:"id"`
			Job string `json:"job"`
		}{ID: j.Options.ID, Job: j.Job},
	)

	return ctx
}

func (j *Job) Ack() {

}

func (j *Job) Nack() {

}
