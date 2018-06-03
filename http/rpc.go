package http

import (
	"github.com/sirupsen/logrus"
)

type RPCServer struct {
	Service *Service
}

// WorkerList contains list of workers.
type WorkerList struct {
	// Workers is list of workers.
	Workers []Worker `json:"workers"`
}

// Worker provides information about specific worker.
type Worker struct {
	// Pid contains process id.
	Pid int `json:"pid"`

	// Status of the worker.
	Status string `json:"status"`

	// Number of worker executions.
	NumExecs uint64 `json:"numExecs"`

	// Created is unix nano timestamp of worker creation time.
	Created int64 `json:"created"`

	// Updated is unix nano timestamp of last worker execution.
	Updated int64 `json:"updated"`
}

// Reset resets underlying RR worker pool and restarts all of it's workers.
func (rpc *RPCServer) Reset(reset bool, r *string) error {
	logrus.Info("resetting worker pool")
	*r = "OK"

	return rpc.Service.srv.rr.Reset()
}

// Workers returns list of active workers and their stats.
func (rpc *RPCServer) Workers(list bool, r *WorkerList) error {
	for _, w := range rpc.Service.srv.rr.Workers() {
		state := w.State()
		r.Workers = append(r.Workers, Worker{
			Pid:      *w.Pid,
			Status:   state.String(),
			NumExecs: state.NumExecs(),
			Created:  w.Created.UnixNano(),
			Updated:  state.Updated().UnixNano(),
		})
	}

	return nil
}
