package jobs

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	jobsv1beta "github.com/spiral/roadrunner/v2/proto/jobs/v1beta"
)

type rpc struct {
	log logger.Logger
	p   *Plugin
}

/*
List of the RPC methods:
1. Push - single job push
2. PushBatch - push job batch

3. Reset - managed by the Resetter plugin

4. Pause - pauses set of pipelines
5. Resume - resumes set of pipelines

6. Workers - managed by the Informer plugin.
7. Stat - jobs statistic
*/

func (r *rpc) Push(j *jobsv1beta.PushRequest, _ *jobsv1beta.Empty) error {
	const op = errors.Op("jobs_rpc_push")

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
	const op = errors.Op("jobs_rpc_push")

	l := len(j.GetJobs())

	batch := make([]*structs.Job, l)

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

func (r *rpc) Pause(req *jobsv1beta.MaintenanceRequest, _ *jobsv1beta.Empty) error {
	pipelines := make([]string, len(req.GetPipelines()))

	for i := 0; i < len(pipelines); i++ {
		pipelines[i] = req.GetPipelines()[i]
	}

	r.p.Pause(pipelines)
	return nil
}

func (r *rpc) Resume(req *jobsv1beta.MaintenanceRequest, _ *jobsv1beta.Empty) error {
	pipelines := make([]string, len(req.GetPipelines()))

	for i := 0; i < len(pipelines); i++ {
		pipelines[i] = req.GetPipelines()[i]
	}

	r.p.Resume(pipelines)
	return nil
}

func (r *rpc) List(_ *jobsv1beta.Empty, resp *jobsv1beta.List) error {
	resp.Pipelines = r.p.List()
	return nil
}

// from converts from transport entity to domain
func (r *rpc) from(j *jobsv1beta.Job) *structs.Job {
	headers := map[string][]string{}

	for k, v := range j.GetHeaders() {
		headers[k] = v.GetValue()
	}

	jb := &structs.Job{
		Job:     j.GetJob(),
		Headers: headers,
		Ident:   j.GetId(),
		Payload: j.GetPayload(),
		Options: &structs.Options{
			Priority:   j.GetOptions().GetPriority(),
			Pipeline:   j.GetOptions().GetPipeline(),
			Delay:      j.GetOptions().GetDelay(),
			Attempts:   j.GetOptions().GetAttempts(),
			RetryDelay: j.GetOptions().GetRetryDelay(),
			Timeout:    j.GetOptions().GetTimeout(),
		},
	}

	return jb
}
