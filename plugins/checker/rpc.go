package checker

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/interfaces/log"
	"github.com/spiral/roadrunner/v2/interfaces/status"
)

type rpc struct {
	srv *Plugin
	log log.Logger
}

// Status return current status of the provided plugin
func (rpc *rpc) Status(service string, status *status.Status) error {
	const op = errors.Op("status")
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
