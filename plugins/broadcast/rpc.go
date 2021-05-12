package broadcast

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type rpc struct {
	log logger.Logger
	svc *Plugin
}

func (r *rpc) Publish(msg []*Message, ok *bool) error {
	const op = errors.Op("broadcast_publish")
	err := r.svc.Publish(msg)
	if err != nil {
		*ok = false
		return errors.E(op, err)
	}
	*ok = true
	return nil
}

func (r *rpc) PublishAsync() {

}
