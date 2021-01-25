package protocol

import (
	v1 "github.com/golang/protobuf/proto" //nolint:staticcheck
	jsoniter "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/payload"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/temporal/protocol/internal"
	"google.golang.org/protobuf/proto"
)

type (
	// ProtoCodec uses protobuf to exchange messages with underlying workers.
	ProtoCodec struct {
	}
)

// NewProtoCodec creates new Proto communication codec.
func NewProtoCodec() Codec {
	return &ProtoCodec{}
}

// WithLogger creates new codes instance with attached logger.
func (c *ProtoCodec) WithLogger(logger logger.Logger) Codec {
	return &ProtoCodec{}
}

// GetName returns codec name.
func (c *ProtoCodec) GetName() string {
	return "protobuf"
}

// Execute exchanges commands with worker.
func (c *ProtoCodec) Execute(e Endpoint, ctx Context, msg ...Message) ([]Message, error) {
	if len(msg) == 0 {
		return nil, nil
	}

	var request = &internal.Frame{}
	var response = &internal.Frame{}
	var result = make([]Message, 0, 5)
	var err error

	for _, m := range msg {
		frame, err := c.packMessage(m)
		if err != nil {
			return nil, err
		}

		request.Messages = append(request.Messages, frame)
	}

	p := payload.Payload{}

	// context is always in json format
	if ctx.IsEmpty() {
		p.Context = []byte("null")
	}

	p.Context, err = jsoniter.Marshal(ctx)
	if err != nil {
		return nil, errors.E(errors.Op("encodeContext"), err)
	}

	p.Body, err = proto.Marshal(v1.MessageV2(request))
	if err != nil {
		return nil, errors.E(errors.Op("encodePayload"), err)
	}

	out, err := e.Exec(p)
	if err != nil {
		return nil, errors.E(errors.Op("execute"), err)
	}

	if len(out.Body) == 0 {
		// worker inactive or closed
		return nil, nil
	}

	err = proto.Unmarshal(out.Body, v1.MessageV2(response))
	if err != nil {
		return nil, errors.E(errors.Op("parseResponse"), err)
	}

	for _, f := range response.Messages {
		msg, err := c.parseMessage(f)
		if err != nil {
			return nil, err
		}

		result = append(result, msg)
	}

	return result, nil
}

func (c *ProtoCodec) packMessage(msg Message) (*internal.Message, error) {
	var err error

	frame := &internal.Message{
		Id:       msg.ID,
		Payloads: msg.Payloads,
		Failure:  msg.Failure,
	}

	if msg.Command != nil {
		frame.Command, err = commandName(msg.Command)
		if err != nil {
			return nil, err
		}

		frame.Options, err = jsoniter.Marshal(msg.Command)
		if err != nil {
			return nil, err
		}
	}

	return frame, nil
}

func (c *ProtoCodec) parseMessage(frame *internal.Message) (Message, error) {
	const op = errors.Op("proto_codec_parse_message")
	var err error

	msg := Message{
		ID:       frame.Id,
		Payloads: frame.Payloads,
		Failure:  frame.Failure,
	}

	if frame.Command != "" {
		msg.Command, err = initCommand(frame.Command)
		if err != nil {
			return Message{}, errors.E(op, err)
		}

		err = jsoniter.Unmarshal(frame.Options, &msg.Command)
		if err != nil {
			return Message{}, errors.E(op, err)
		}
	}

	return msg, nil
}
