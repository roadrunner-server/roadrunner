package tools

import (
	"github.com/shirou/gopsutil/process"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/worker"
)

// ProcessState provides information about specific worker.
type ProcessState struct {
	// Pid contains process id.
	Pid int `json:"pid"`

	// Status of the worker.
	Status string `json:"status"`

	// Number of worker executions.
	NumJobs uint64 `json:"numExecs"`

	// Created is unix nano timestamp of worker creation time.
	Created int64 `json:"created"`

	// MemoryUsage holds the information about worker memory usage in bytes.
	// Values might vary for different operating systems and based on RSS.
	MemoryUsage uint64 `json:"memoryUsage"`
}

// WorkerProcessState creates new worker state definition.
func WorkerProcessState(w worker.BaseProcess) (ProcessState, error) {
	const op = errors.Op("worker_process_state")
	p, _ := process.NewProcess(int32(w.Pid()))
	i, err := p.MemoryInfo()
	if err != nil {
		return ProcessState{}, errors.E(op, err)
	}

	return ProcessState{
		Pid:         int(w.Pid()),
		Status:      w.State().String(),
		NumJobs:     w.State().NumExecs(),
		Created:     w.Created().UnixNano(),
		MemoryUsage: i.RSS,
	}, nil
}
