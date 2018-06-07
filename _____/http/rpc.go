package http

import (
	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner/_____/utils"
	"github.com/pkg/errors"
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
	if rpc.service.srv == nil {
		return errors.New("no http server")
	}

	logrus.Info("http: restarting worker pool")
	*r = "OK"

	err := rpc.service.srv.rr.Reset()
	if err != nil {
		logrus.Errorf("http: %s", err)
	}

	return err
}

// Workers returns list of active workers and their stats.
func (rpc *rpcServer) Workers(list bool, r *WorkerList) error {
	if rpc.service.srv == nil {
		return errors.New("no http server")
	}

	r.Workers = utils.FetchWorkers(rpc.service.srv.rr)
	return nil
}
