package jobs

import (
	"context"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/jobs/job"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	jobsv1beta "github.com/spiral/roadrunner/v2/proto/jobs/v1beta"
)

type rpc struct {
	log logger.Logger
	p   *Plugin
}

func (r *rpc) Push(j *jobsv1beta.PushRequest, _ *jobsv1beta.Empty) error {
	const op = errors.Op("rpc_push")

	// convert transport entity into domain
	// how we can do this quickly

	if j.GetJob().GetId() == "" {
		return errors.E(op, errors.Str("empty ID field not allowed"))
	}

	err := r.p.Push(r.from(j.GetJob()))
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (r *rpc) PushBatch(j *jobsv1beta.PushBatchRequest, _ *jobsv1beta.Empty) error {
	const op = errors.Op("rpc_push_batch")

	l := len(j.GetJobs())

	batch := make([]*job.Job, l)

	for i := 0; i < l; i++ {
		// convert transport entity into domain
		// how we can do this quickly
		batch[i] = r.from(j.GetJobs()[i])
	}

	err := r.p.PushBatch(batch)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (r *rpc) Pause(req *jobsv1beta.Pipelines, _ *jobsv1beta.Empty) error {
	for i := 0; i < len(req.GetPipelines()); i++ {
		r.p.Pause(req.GetPipelines()[i])
	}

	return nil
}

func (r *rpc) Resume(req *jobsv1beta.Pipelines, _ *jobsv1beta.Empty) error {
	for i := 0; i < len(req.GetPipelines()); i++ {
		r.p.Resume(req.GetPipelines()[i])
	}

	return nil
}

func (r *rpc) List(_ *jobsv1beta.Empty, resp *jobsv1beta.Pipelines) error {
	resp.Pipelines = r.p.List()
	return nil
}

// Declare pipeline used to dynamically declare any type of the pipeline
// Mandatory fields:
// 1. Driver
// 2. Pipeline name
// 3. Options related to the particular pipeline
func (r *rpc) Declare(req *jobsv1beta.DeclareRequest, _ *jobsv1beta.Empty) error {
	const op = errors.Op("rpc_declare_pipeline")
	pipe := &pipeline.Pipeline{}

	for i := range req.GetPipeline() {
		(*pipe)[i] = req.GetPipeline()[i]
	}

	err := r.p.Declare(pipe)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (r *rpc) Destroy(req *jobsv1beta.Pipelines, resp *jobsv1beta.Pipelines) error {
	const op = errors.Op("rpc_declare_pipeline")

	var destroyed []string //nolint:prealloc
	for i := 0; i < len(req.GetPipelines()); i++ {
		err := r.p.Destroy(req.GetPipelines()[i])
		if err != nil {
			return errors.E(op, err)
		}
		destroyed = append(destroyed, req.GetPipelines()[i])
	}

	// return destroyed pipelines
	resp.Pipelines = destroyed

	return nil
}

func (r *rpc) Stat(_ *jobsv1beta.Empty, resp *jobsv1beta.Stats) error {
	const op = errors.Op("rpc_stats")
	state, err := r.p.JobsState(context.Background())
	if err != nil {
		return errors.E(op, err)
	}

	for i := 0; i < len(state); i++ {
		resp.Stats = append(resp.Stats, &jobsv1beta.Stat{
			Pipeline: state[i].Pipeline,
			Driver:   state[i].Driver,
			Queue:    state[i].Queue,
			Active:   state[i].Active,
			Delayed:  state[i].Delayed,
			Reserved: state[i].Reserved,
		})
	}

	return nil
}

// from converts from transport entity to domain
func (r *rpc) from(j *jobsv1beta.Job) *job.Job {
	headers := map[string][]string{}

	for k, v := range j.GetHeaders() {
		headers[k] = v.GetValue()
	}

	jb := &job.Job{
		Job:     j.GetJob(),
		Headers: headers,
		Ident:   j.GetId(),
		Payload: j.GetPayload(),
		Options: &job.Options{
			Priority: j.GetOptions().GetPriority(),
			Pipeline: j.GetOptions().GetPipeline(),
			Delay:    j.GetOptions().GetDelay(),
		},
	}

	return jb
}
