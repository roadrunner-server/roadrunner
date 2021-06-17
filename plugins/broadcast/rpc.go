package broadcast

import (
	"github.com/spiral/errors"
	websocketsv1 "github.com/spiral/roadrunner/v2/pkg/proto/websockets/v1beta"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"google.golang.org/protobuf/proto"
)

// rpc collectors struct
type rpc struct {
	plugin *Plugin
	log    logger.Logger
}

// Publish ... msg is a proto decoded payload
// see: pkg/pubsub/message.fbs
func (r *rpc) Publish(in *websocketsv1.Request, out *websocketsv1.Response) error {
	const op = errors.Op("broadcast_publish")

	// just return in case of nil message
	if in == nil {
		out.Ok = false
		return nil
	}

	r.log.Debug("message published", "msg", in.Messages)

	msgLen := len(in.GetMessages())

	for i := 0; i < msgLen; i++ {
		bb, err := proto.Marshal(in.GetMessages()[i])
		if err != nil {
			return errors.E(op, err)
		}

		err = r.plugin.Publish(bb)
		if err != nil {
			out.Ok = false
			return errors.E(op, err)
		}
	}

	out.Ok = true
	return nil
}

// PublishAsync ...
// see: pkg/pubsub/message.fbs
func (r *rpc) PublishAsync(in *websocketsv1.Request, out *websocketsv1.Response) error {
	const op = errors.Op("publish_async")

	// just return in case of nil message
	if in == nil {
		out.Ok = false
		return nil
	}

	r.log.Debug("message published", "msg", in.Messages)

	msgLen := len(in.GetMessages())

	for i := 0; i < msgLen; i++ {
		bb, err := proto.Marshal(in.GetMessages()[i])
		if err != nil {
			out.Ok = false
			return errors.E(op, err)
		}

		r.plugin.PublishAsync(bb)
	}

	out.Ok = true
	return nil
}
