package process

import (
	"github.com/shirou/gopsutil/process"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/worker"
)

// State provides information about specific worker.
type State struct {
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

	// CPU_Percent returns how many percent of the CPU time this process uses
	CPUPercent float64

	// Command used in the service plugin and shows a command for the particular service
	Command string
}

// WorkerProcessState creates new worker state definition.
func WorkerProcessState(w worker.BaseProcess) (State, error) {
	const op = errors.Op("worker_process_state")
	p, _ := process.NewProcess(int32(w.Pid()))
	i, err := p.MemoryInfo()
	if err != nil {
		return State{}, errors.E(op, err)
	}

	percent, err := p.CPUPercent()
	if err != nil {
		return State{}, err
	}

	return State{
		CPUPercent:  percent,
		Pid:         int(w.Pid()),
		Status:      w.State().String(),
		NumJobs:     w.State().NumExecs(),
		Created:     w.Created().UnixNano(),
		MemoryUsage: i.RSS,
	}, nil
}

func GeneralProcessState(pid int, command string) (State, error) {
	const op = errors.Op("process_state")
	p, _ := process.NewProcess(int32(pid))
	i, err := p.MemoryInfo()
	if err != nil {
		return State{}, errors.E(op, err)
	}
	percent, err := p.CPUPercent()
	if err != nil {
		return State{}, err
	}

	return State{
		CPUPercent:  percent,
		Pid:         pid,
		MemoryUsage: i.RSS,
		Command:     command,
	}, nil
}
