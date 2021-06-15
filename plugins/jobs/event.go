package jobs

import "time"

const (
	// EventPushOK thrown when new job has been added. JobEvent is passed as context.
	EventPushOK = iota + 1500

	// EventPushError caused when job can not be registered.
	EventPushError

	// EventJobStart thrown when new job received.
	EventJobStart

	// EventJobOK thrown when job execution is successfully completed. JobEvent is passed as context.
	EventJobOK

	// EventJobError thrown on all job related errors. See JobError as context.
	EventJobError

	// EventPipeConsume when pipeline pipelines has been requested.
	EventPipeConsume

	// EventPipeActive when pipeline has started.
	EventPipeActive

	// EventPipeStop when pipeline has begun stopping.
	EventPipeStop

	// EventPipeStopped when pipeline has been stopped.
	EventPipeStopped

	// EventPipeError when pipeline specific error happen.
	EventPipeError

	// EventBrokerReady thrown when broken is ready to accept/serve tasks.
	EventBrokerReady
)

// JobEvent represent job event.
type JobEvent struct {
	// String is job id.
	ID string

	// Job is failed job.
	Job *Job

	// event timings
	start   time.Time
	elapsed time.Duration
}

// Elapsed returns duration of the invocation.
func (e *JobEvent) Elapsed() time.Duration {
	return e.elapsed
}

// JobError represents singular Job error event.
type JobError struct {
	// String is job id.
	ID string

	// Job is failed job.
	Job *Job

	// Caused contains job specific error.
	Caused error

	// event timings
	start   time.Time
	elapsed time.Duration
}

// Elapsed returns duration of the invocation.
func (e *JobError) Elapsed() time.Duration {
	return e.elapsed
}

// Caused returns error message.
func (e *JobError) Error() string {
	return e.Caused.Error()
}

// PipelineError defines pipeline specific errors.
type PipelineError struct {
	// Pipeline is associated pipeline.
	Pipeline *Pipeline

	// Caused send by broker.
	Caused error
}

// Error returns error message.
func (e *PipelineError) Error() string {
	return e.Caused.Error()
}
