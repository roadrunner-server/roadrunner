package oooold

import (
	"fmt"
)

type rpcServer struct{ svc *Service }

// WorkerList contains list of workers.
type WorkerList struct {
	// Workers is list of workers.
	Workers []*util.State `json:"workers"`
}

// PipelineList contains list of pipeline stats.
type PipelineList struct {
	// Pipelines is list of pipeline stats.
	Pipelines []*Stat `json:"pipelines"`
}

// Push job to the testQueue.
func (rpc *rpcServer) Push(j *Job, id *string) (err error) {
	if rpc.svc == nil {
		return fmt.Errorf("jobs server is not running")
	}

	*id, err = rpc.svc.Push(j)
	return
}

// Push job to the testQueue.
func (rpc *rpcServer) PushAsync(j *Job, ok *bool) (err error) {
	if rpc.svc == nil {
		return fmt.Errorf("jobs server is not running")
	}

	*ok = true
	go rpc.svc.Push(j)

	return
}

// Reset resets underlying RR worker pool and restarts all of it's workers.
func (rpc *rpcServer) Reset(reset bool, w *string) error {
	if rpc.svc == nil {
		return fmt.Errorf("jobs server is not running")
	}

	*w = "OK"
	return rpc.svc.rr.Reset()
}

// Destroy job pipelines for a given pipeline.
func (rpc *rpcServer) Stop(pipeline string, w *string) (err error) {
	if rpc.svc == nil {
		return fmt.Errorf("jobs server is not running")
	}

	pipe := rpc.svc.cfg.pipelines.Get(pipeline)
	if pipe == nil {
		return fmt.Errorf("undefined pipeline `%s`", pipeline)
	}

	if err := rpc.svc.Consume(pipe, nil, nil); err != nil {
		return err
	}

	*w = "OK"
	return nil
}

// Resume job pipelines for a given pipeline.
func (rpc *rpcServer) Resume(pipeline string, w *string) (err error) {
	if rpc.svc == nil {
		return fmt.Errorf("jobs server is not running")
	}

	pipe := rpc.svc.cfg.pipelines.Get(pipeline)
	if pipe == nil {
		return fmt.Errorf("undefined pipeline `%s`", pipeline)
	}

	if err := rpc.svc.Consume(pipe, rpc.svc.execPool, rpc.svc.error); err != nil {
		return err
	}

	*w = "OK"
	return nil
}

// Destroy job pipelines for a given pipeline.
func (rpc *rpcServer) StopAll(stop bool, w *string) (err error) {
	if rpc.svc == nil || rpc.svc.rr == nil {
		return fmt.Errorf("jobs server is not running")
	}

	for _, pipe := range rpc.svc.cfg.pipelines {
		if err := rpc.svc.Consume(pipe, nil, nil); err != nil {
			return err
		}
	}

	*w = "OK"
	return nil
}

// Resume job pipelines for a given pipeline.
func (rpc *rpcServer) ResumeAll(resume bool, w *string) (err error) {
	if rpc.svc == nil {
		return fmt.Errorf("jobs server is not running")
	}

	for _, pipe := range rpc.svc.cfg.pipelines {
		if err := rpc.svc.Consume(pipe, rpc.svc.execPool, rpc.svc.error); err != nil {
			return err
		}
	}

	*w = "OK"
	return nil
}

// Workers returns list of pipelines workers and their stats.
func (rpc *rpcServer) Workers(list bool, w *WorkerList) (err error) {
	if rpc.svc == nil {
		return fmt.Errorf("jobs server is not running")
	}

	w.Workers, err = util.ServerState(rpc.svc.rr)
	return err
}

// Stat returns list of pipelines workers and their stats.
func (rpc *rpcServer) Stat(list bool, l *PipelineList) (err error) {
	if rpc.svc == nil {
		return fmt.Errorf("jobs server is not running")
	}

	*l = PipelineList{}
	for _, p := range rpc.svc.cfg.pipelines {
		stat, err := rpc.svc.Stat(p)
		if err != nil {
			return err
		}

		l.Pipelines = append(l.Pipelines, stat)
	}

	return err
}
