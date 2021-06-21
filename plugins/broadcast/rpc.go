package broadcast

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	websocketsv1 "github.com/spiral/roadrunner/v2/proto/websockets/v1beta"
)

// rpc collectors struct
type rpc struct {
	plugin *Plugin
	log    logger.Logger
}

// Publish ... msg is a proto decoded payload
// see: root/proto
func (r *rpc) Publish(in *websocketsv1.Request, out *websocketsv1.Response) error {
	const op = errors.Op("broadcast_publish")

	// just return in case of nil message
	if in == nil {
		out.Ok = false
		return nil
	}

	r.log.Debug("message published", "msg", in.String())
	msgLen := len(in.GetMessages())

	for i := 0; i < msgLen; i++ {
		for j := 0; j < len(in.GetMessages()[i].GetTopics()); j++ {
			if in.GetMessages()[i].GetTopics()[j] == "" {
				r.log.Warn("message with empty topic, skipping")
				// skip empty topics
				continue
			}

			tmp := &pubsub.Message{
				Topic:   in.GetMessages()[i].GetTopics()[j],
				Payload: in.GetMessages()[i].GetPayload(),
			}

			err := r.plugin.Publish(tmp)
			if err != nil {
				out.Ok = false
				return errors.E(op, err)
			}
		}
	}

	out.Ok = true
	return nil
}

// PublishAsync ...
// see: root/proto
func (r *rpc) PublishAsync(in *websocketsv1.Request, out *websocketsv1.Response) error {
	// just return in case of nil message
	if in == nil {
		out.Ok = false
		return nil
	}

	r.log.Debug("message published", "msg", in.GetMessages())

	msgLen := len(in.GetMessages())

	for i := 0; i < msgLen; i++ {
		for j := 0; j < len(in.GetMessages()[i].GetTopics()); j++ {
			if in.GetMessages()[i].GetTopics()[j] == "" {
				r.log.Warn("message with empty topic, skipping")
				// skip empty topics
				continue
			}

			tmp := &pubsub.Message{
				Topic:   in.GetMessages()[i].GetTopics()[j],
				Payload: in.GetMessages()[i].GetPayload(),
			}

			r.plugin.PublishAsync(tmp)
		}
	}

	out.Ok = true
	return nil
}
