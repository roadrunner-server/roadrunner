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

4. Stop - stop pipeline processing
5. StopAll - stop all pipelines processing
6. Resume - resume pipeline processing
7. ResumeAll - resume stopped pipelines

8. Workers - managed by the Informer plugin.
9. Stat - jobs statistic
*/

func (r *rpc) Push(j *jobsv1beta.Request, resp *jobsv1beta.Response) error {
	const op = errors.Op("jobs_rpc_push")

	// convert transport entity into domain
	// how we can do this quickly
	jb := &structs.Job{
		Job:     j.GetJob().Job,
		Payload: j.GetJob().Payload,
		Options: &structs.Options{
			Priority:   &j.GetJob().Options.Priority,
			ID:         &j.GetJob().Options.Id,
			Pipeline:   j.GetJob().Options.Pipeline,
			Delay:      j.GetJob().Options.Delay,
			Attempts:   j.GetJob().Options.Attempts,
			RetryDelay: j.GetJob().Options.RetryDelay,
			Timeout:    j.GetJob().Options.Timeout,
		},
	}

	id, err := r.p.Push(jb)
	if err != nil {
		return errors.E(op, err)
	}

	resp.Id = *id

	return nil
}

func (r *rpc) PushBatch(j *jobsv1beta.BatchRequest, resp *jobsv1beta.Response) error {
	const op = errors.Op("jobs_rpc_push")

	l := len(j.GetJobs())

	batch := make([]*structs.Job, l)

	for i := 0; i < l; i++ {
		// convert transport entity into domain
		// how we can do this quickly
		jb := &structs.Job{
			Job:     j.GetJobs()[i].Job,
			Payload: j.GetJobs()[i].Payload,
			Options: &structs.Options{
				Priority:   &j.GetJobs()[i].Options.Priority,
				ID:         &j.GetJobs()[i].Options.Id,
				Pipeline:   j.GetJobs()[i].Options.Pipeline,
				Delay:      j.GetJobs()[i].Options.Delay,
				Attempts:   j.GetJobs()[i].Options.Attempts,
				RetryDelay: j.GetJobs()[i].Options.RetryDelay,
				Timeout:    j.GetJobs()[i].Options.Timeout,
			},
		}

		batch[i] = jb
	}

	_, err := r.p.PushBatch(batch)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (r *rpc) Stop(pipeline string, w *string) error {
	return nil
}

func (r *rpc) StopAll(_ bool, w *string) error {
	return nil
}

func (r *rpc) Resume(pipeline string, w *string) error {
	return nil
}

func (r *rpc) ResumeAll(_ bool, w *string) error {
	return nil
}
