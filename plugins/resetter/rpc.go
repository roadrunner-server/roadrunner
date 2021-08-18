package resetter

import "github.com/spiral/errors"

type rpc struct {
	srv *Plugin
}

// List all resettable plugins.
func (rpc *rpc) List(_ bool, list *[]string) error {
	*list = make([]string, 0)

	for name := range rpc.srv.registry {
		*list = append(*list, name)
	}
	return nil
}

// Reset named plugin.
func (rpc *rpc) Reset(service string, done *bool) error {
	const op = errors.Op("resetter_rpc_reset")
	err := rpc.srv.Reset(service)
	if err != nil {
		*done = false
		return errors.E(op, err)
	}
	*done = true
	return nil
}
