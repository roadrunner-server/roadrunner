package checker

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type rpc struct {
	srv *Plugin
	log logger.Logger
}

// Status return current status of the provided plugin
func (rpc *rpc) Status(service string, status *Status) error {
	const op = errors.Op("checker_rpc_status")
	rpc.log.Debug("started Status method", "service", service)
	st, err := rpc.srv.Status(service)
	if err != nil {
		return errors.E(op, err)
	}

	*status = st

	rpc.log.Debug("status code", "code", st.Code)
	rpc.log.Debug("successfully finished Status method")
	return nil
}
