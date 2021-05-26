package websockets

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type rpc struct {
	plugin *Plugin
	log    logger.Logger
}

func (r *rpc) Publish(msg []*pubsub.Msg, ok *bool) error {
	const op = errors.Op("broadcast_publish")

	// publish to the registered broker
	mi := make([]pubsub.Message, 0, len(msg))
	// golang can't convert slice in-place
	// so, we need to convert it manually
	for i := 0; i < len(msg); i++ {
		mi = append(mi, msg[i])
	}
	err := r.plugin.Publish(mi)
	if err != nil {
		*ok = false
		return errors.E(op, err)
	}
	*ok = true
	return nil
}

func (r *rpc) PublishAsync(msg []*pubsub.Msg, ok *bool) error {
	// publish to the registered broker
	mi := make([]pubsub.Message, 0, len(msg))
	// golang can't convert slice in-place
	// so, we need to convert it manually
	for i := 0; i < len(msg); i++ {
		mi = append(mi, msg[i])
	}

	r.plugin.PublishAsync(mi)

	*ok = true
	return nil
}
