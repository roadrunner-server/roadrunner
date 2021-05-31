package websockets

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

// rpc collectors struct
type rpc struct {
	plugin *Plugin
	log    logger.Logger
}

func (r *rpc) Publish(msg []*pubsub.Message, ok *bool) error {
	const op = errors.Op("broadcast_publish")
	r.log.Debug("message published", "msg", msg)

	// just return in case of nil message
	if msg == nil {
		*ok = true
		return nil
	}

	err := r.plugin.Publish(msg)
	if err != nil {
		*ok = false
		return errors.E(op, err)
	}
	*ok = true
	return nil
}

func (r *rpc) PublishAsync(msg []*pubsub.Message, ok *bool) error {
	r.log.Debug("message published", "msg", msg)

	// just return in case of nil message
	if msg == nil {
		*ok = true
		return nil
	}
	// publish to the registered broker
	r.plugin.PublishAsync(msg)

	*ok = true
	return nil
}
