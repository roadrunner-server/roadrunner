package jobs

import (
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type rpc struct {
	log logger.Logger
	p   *Plugin
}

func (r *rpc) Push(j *structs.Job, idRet *string) error {
	id, err := r.p.Push(j)
	if err != nil {
		panic(err)
	}
	*idRet = id
	return nil
}
