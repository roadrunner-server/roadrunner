package util

import (
	"errors"
	"github.com/shirou/gopsutil/process"
	"github.com/spiral/roadrunner"
)

// State provides information about specific worker.
type State struct {
	// Pid contains process id.
	Pid int `json:"pid"`

	// Status of the worker.
	Status string `json:"status"`

	// Number of worker executions.
	NumJobs int64 `json:"numExecs"`

	// Created is unix nano timestamp of worker creation time.
	Created int64 `json:"created"`

	// MemoryUsage holds the information about worker memory usage in bytes.
	// Values might vary for different operating systems and based on RSS.
	MemoryUsage uint64 `json:"memoryUsage"`
}

// WorkerState creates new worker state definition.
func WorkerState(w *roadrunner.Worker) (*State, error) {
	p, _ := process.NewProcess(int32(*w.Pid))
	i, err := p.MemoryInfo()
	if err != nil {
		return nil, err
	}

	return &State{
		Pid:         *w.Pid,
		Status:      w.State().String(),
		NumJobs:     w.State().NumExecs(),
		Created:     w.Created.UnixNano(),
		MemoryUsage: i.RSS,
	}, nil
}

// ServerState returns list of all worker states of a given rr server.
func ServerState(rr *roadrunner.Server) ([]*State, error) {
	if rr == nil {
		return nil, errors.New("rr server is not running")
	}

	result := make([]*State, 0)
	for _, w := range rr.Workers() {
		state, err := WorkerState(w)
		if err != nil {
			return nil, err
		}

		result = append(result, state)
	}

	return result, nil
}
