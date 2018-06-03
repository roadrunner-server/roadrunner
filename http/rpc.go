package http

import (
	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner/utils"
)

type rpcServer struct {
	service *Service
}

// WorkerList contains list of workers.
type WorkerList struct {
	// Workers is list of workers.
	Workers []utils.Worker `json:"workers"`
}

// Reset resets underlying RR worker pool and restarts all of it's workers.
func (rpc *rpcServer) Reset(reset bool, r *string) error {
	logrus.Info("http: restarting worker pool")
	*r = "OK"

	return rpc.service.srv.rr.Reset()
}

// Workers returns list of active workers and their stats.
func (rpc *rpcServer) Workers(list bool, r *WorkerList) error {
	r.Workers = utils.FetchWorkers(rpc.service.srv.rr)
	return nil
}
