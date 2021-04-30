package informer

import (
	"github.com/spiral/roadrunner/v2/pkg/process"
)

type rpc struct {
	srv *Plugin
}

// WorkerList contains list of workers.
type WorkerList struct {
	// Workers is list of workers.
	Workers []process.State `json:"workers"`
}

// List all resettable services.
func (rpc *rpc) List(_ bool, list *[]string) error {
	*list = make([]string, 0, len(rpc.srv.registry))

	// append all plugin names to the output result
	for name := range rpc.srv.available {
		*list = append(*list, name)
	}
	return nil
}

// Workers state of a given service.
func (rpc *rpc) Workers(service string, list *WorkerList) error {
	workers, err := rpc.srv.Workers(service)
	if err != nil {
		return err
	}

	// write actual processes
	list.Workers = workers

	return nil
}
