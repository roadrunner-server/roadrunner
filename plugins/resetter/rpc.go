package resetter

import "github.com/spiral/roadrunner/v2/plugins/logger"

type rpc struct {
	srv *Plugin
	log logger.Logger
}

// List all resettable plugins.
func (rpc *rpc) List(_ bool, list *[]string) error {
	rpc.log.Debug("started List method")
	*list = make([]string, 0)

	for name := range rpc.srv.registry {
		*list = append(*list, name)
	}
	rpc.log.Debug("services list", "services", *list)

	rpc.log.Debug("finished List method")
	return nil
}

// Reset named plugin.
func (rpc *rpc) Reset(service string, done *bool) error {
	rpc.log.Debug("started Reset method for the service", "service", service)
	defer rpc.log.Debug("finished Reset method for the service", "service", service)
	*done = true
	return rpc.srv.Reset(service)
}
