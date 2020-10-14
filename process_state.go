package roadrunner

import (
	"github.com/shirou/gopsutil/process"
)

// ProcessState provides information about specific worker.
type ProcessState struct {
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

// WorkerProcessState creates new worker state definition.
func WorkerProcessState(w WorkerBase) (ProcessState, error) {
	p, _ := process.NewProcess(int32(w.Pid()))
	i, err := p.MemoryInfo()
	if err != nil {
		return ProcessState{}, err
	}

	return ProcessState{
		Pid:         int(w.Pid()),
		Status:      w.State().String(),
		NumJobs:     w.State().NumExecs(),
		Created:     w.Created().UnixNano(),
		MemoryUsage: i.RSS,
	}, nil
}

// ServerState returns list of all worker states of a given rr server.
func PoolState(pool Pool) ([]ProcessState, error) {
	result := make([]ProcessState, 0)
	for _, w := range pool.Workers() {
		state, err := WorkerProcessState(w)
		if err != nil {
			return nil, err
		}

		result = append(result, state)
	}

	return result, nil
}
