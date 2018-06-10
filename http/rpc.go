package http

import (
	"github.com/pkg/errors"
)

type rpcServer struct{ svc *Service }

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
	NumJobs uint64 `json:"numExecs"`

	// Created is unix nano timestamp of worker creation time.
	Created int64 `json:"created"`

	// Updated is unix nano timestamp of last worker execution.
	Updated int64 `json:"updated"`
}

// Reset resets underlying RR worker pool and restarts all of it's workers.
func (rpc *rpcServer) Reset(reset bool, r *string) error {
	if rpc.svc.srv == nil {
		return errors.New("http server is not running")
	}

	*r = "OK"
	return rpc.svc.srv.rr.Reset()
}

// Workers returns list of active workers and their stats.
func (rpc *rpcServer) Workers(list bool, r *WorkerList) error {
	if rpc.svc.srv == nil {
		return errors.New("http server is not running")
	}

	for _, w := range rpc.svc.rr.Workers() {
		state := w.State()
		r.Workers = append(r.Workers, Worker{
			Pid:     *w.Pid,
			Status:  state.String(),
			NumJobs: state.NumExecs(),
			Created: w.Created.UnixNano(),
			Updated: state.Updated().UnixNano(),
		})
	}

	return nil
}
