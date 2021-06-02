package websockets

import (
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pubsub/message"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

// rpc collectors struct
type rpc struct {
	plugin *Plugin
	log    logger.Logger
}

// Publish ... msg is a flatbuffers decoded payload
// see: pkg/pubsub/message.fbs
func (r *rpc) Publish(msg []byte, ok *bool) error {
	const op = errors.Op("broadcast_publish")
	r.log.Debug("message published")

	// just return in case of nil message
	if msg == nil {
		*ok = true
		return nil
	}

	fbsMsg := message.GetRootAsMessages(msg, 0)
	tmpMsg := &message.Message{}

	b := flatbuffers.NewBuilder(100)

	for i := 0; i < fbsMsg.MessagesLength(); i++ {
		// init a message
		fbsMsg.Messages(tmpMsg, i)

		// overhead HERE
		orig := serializeMsg(b, tmpMsg)
		bb := make([]byte, len(orig))
		copy(bb, orig)

		err := r.plugin.Publish(bb)
		if err != nil {
			*ok = false
			b.Reset()
			return errors.E(op, err)
		}
		b.Reset()
	}

	*ok = true
	return nil
}

// PublishAsync ...
// see: pkg/pubsub/message.fbs
func (r *rpc) PublishAsync(msg []byte, ok *bool) error {
	r.log.Debug("message published", "msg", msg)

	// just return in case of nil message
	if msg == nil {
		*ok = true
		return nil
	}

	fbsMsg := message.GetRootAsMessages(msg, 0)
	tmpMsg := &message.Message{}

	b := flatbuffers.NewBuilder(100)

	for i := 0; i < fbsMsg.MessagesLength(); i++ {
		// init a message
		fbsMsg.Messages(tmpMsg, i)

		// overhead HERE
		orig := serializeMsg(b, tmpMsg)
		bb := make([]byte, len(orig))
		copy(bb, orig)

		r.plugin.PublishAsync(bb)
		b.Reset()
	}

	*ok = true
	return nil
}

func serializeMsg(b *flatbuffers.Builder, msg *message.Message) []byte {
	cmdOff := b.CreateByteString(msg.Command())
	brokerOff := b.CreateByteString(msg.Broker())

	offsets := make([]flatbuffers.UOffsetT, msg.TopicsLength())
	for j := msg.TopicsLength() - 1; j >= 0; j-- {
		offsets[j] = b.CreateByteString(msg.Topics(j))
	}

	message.MessageStartTopicsVector(b, len(offsets))

	for j := len(offsets) - 1; j >= 0; j-- {
		b.PrependUOffsetT(offsets[j])
	}

	tOff := b.EndVector(len(offsets))
	bb := make([]byte, msg.PayloadLength())
	for i := 0; i < msg.PayloadLength(); i++ {
		bb[i] = byte(msg.Payload(i))
	}
	pOff := b.CreateByteVector(bb)

	message.MessageStart(b)

	message.MessageAddCommand(b, cmdOff)
	message.MessageAddBroker(b, brokerOff)
	message.MessageAddTopics(b, tOff)
	message.MessageAddPayload(b, pOff)

	fOff := message.MessageEnd(b)
	b.Finish(fOff)

	return b.FinishedBytes()
}
