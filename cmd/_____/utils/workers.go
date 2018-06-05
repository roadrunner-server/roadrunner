package utils

import "github.com/spiral/roadrunner"

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

// FetchWorkers fetches list of workers from RR Server.
func FetchWorkers(srv *roadrunner.Server) (result []Worker) {
	for _, w := range srv.Workers() {
		state := w.State()
		result = append(result, Worker{
			Pid:      *w.Pid,
			Status:   state.String(),
			NumExecs: state.NumExecs(),
			Created:  w.Created.UnixNano(),
			Updated:  state.Updated().UnixNano(),
		})
	}

	return
}