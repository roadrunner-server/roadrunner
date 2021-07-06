package jobs

import (
	"sync"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	jobsv1beta "github.com/spiral/roadrunner/v2/proto/jobs/v1beta"
)

type rpc struct {
	log logger.Logger
	p   *Plugin
}

var jobsPool = &sync.Pool{
	New: func() interface{} {
		return &structs.Job{
			Options: &structs.Options{},
		}
	},
}

func pubJob(j *structs.Job) {
	// clear
	j.Job = ""
	j.Payload = ""
	j.Options = &structs.Options{}
	jobsPool.Put(j)
}

func getJob() *structs.Job {
	return jobsPool.Get().(*structs.Job)
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
	jb := getJob()
	defer pubJob(jb)

	jb = &structs.Job{
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

func (r *rpc) PushBatch(j *structs.Job, idRet *string) error {
	const op = errors.Op("jobs_rpc_push")
	id, err := r.p.Push(j)
	if err != nil {
		return errors.E(op, err)
	}

	*idRet = *id
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
