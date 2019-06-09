package http

import (
	"github.com/pkg/errors"
	"github.com/spiral/roadrunner/util"
)

type rpcServer struct{ svc *Service }

type WorkersResponse struct {
	Workers []*util.State `json:"workers"`
}

type StatsResponse struct {
	Stats *ServiceStats `json:"stats"`
}

// Reset resets underlying RR worker pool and restarts all of it's workers.
func (rpc *rpcServer) Reset(reset bool, r *string) error {
	if rpc.svc == nil || rpc.svc.handler == nil {
		return errors.New("http server is not running")
	}

	*r = "OK"
	return rpc.svc.Server().Reset()
}

// Workers returns list of active workers and their stats.
func (rpc *rpcServer) Workers(list bool, r *WorkersResponse) (err error) {
	if rpc.svc == nil || rpc.svc.handler == nil {
		return errors.New("http server is not running")
	}

	r.Workers, err = util.ServerState(rpc.svc.Server())
	return err
}

// Stats return stats of Http Service
func (rpc *rpcServer) Stats(uneeded bool, r *StatsResponse) (err error) {
	if rpc.svc == nil || rpc.svc.handler == nil {
		return errors.New("http server is not running")
	}

	r.Stats = rpc.svc.stats;
	return nil;
}
