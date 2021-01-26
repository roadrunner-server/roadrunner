package protocol

import (
	"github.com/fatih/color"
	j "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/payload"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/failure/v1"
)

var json = j.ConfigCompatibleWithStandardLibrary

// JSONCodec can be used for debugging and log capturing reasons.
type JSONCodec struct {
	// level enables verbose logging or all incoming and outcoming messages.
	level DebugLevel

	// logger renders messages when debug enabled.
	logger logger.Logger
}

// jsonFrame contains message command in binary form.
type jsonFrame struct {
	// ID contains ID of the command, response or error.
	ID uint64 `json:"id"`

	// Command name. Optional.
	Command string `json:"command,omitempty"`

	// Options to be unmarshalled to body (raw payload).
	Options j.RawMessage `json:"options,omitempty"`

	// Failure associated with command id.
	Failure []byte `json:"failure,omitempty"`

	// Payloads specific to the command or result.
	Payloads []byte `json:"payloads,omitempty"`
}

// NewJSONCodec creates new Json communication codec.
func NewJSONCodec(level DebugLevel, logger logger.Logger) Codec {
	return &JSONCodec{
		level:  level,
		logger: logger,
	}
}

// WithLogger creates new codes instance with attached logger.
func (c *JSONCodec) WithLogger(logger logger.Logger) Codec {
	return &JSONCodec{
		level:  c.level,
		logger: logger,
	}
}

// GetName returns codec name.
func (c *JSONCodec) GetName() string {
	return "json"
}

// Execute exchanges commands with worker.
func (c *JSONCodec) Execute(e Endpoint, ctx Context, msg ...Message) ([]Message, error) {
	const op = errors.Op("json_codec_execute")
	if len(msg) == 0 {
		return nil, nil
	}

	var response = make([]jsonFrame, 0, 5)
	var result = make([]Message, 0, 5)
	var err error

	frames := make([]jsonFrame, 0, len(msg))
	for _, m := range msg {
		frame, err := c.packFrame(m)
		if err != nil {
			return nil, errors.E(op, err)
		}

		frames = append(frames, frame)
	}

	p := payload.Payload{}

	if ctx.IsEmpty() {
		p.Context = []byte("null")
	}

	p.Context, err = json.Marshal(ctx)
	if err != nil {
		return nil, errors.E(op, err)
	}

	p.Body, err = json.Marshal(frames)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if c.level >= DebugNormal {
		logMessage := string(p.Body) + " " + string(p.Context)
		if c.level >= DebugHumanized {
			logMessage = color.GreenString(logMessage)
		}

		c.logger.Debug(logMessage)
	}

	out, err := e.Exec(p)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if len(out.Body) == 0 {
		// worker inactive or closed
		return nil, nil
	}

	if c.level >= DebugNormal {
		logMessage := string(out.Body)
		if c.level >= DebugHumanized {
			logMessage = color.HiYellowString(logMessage)
		}

		c.logger.Debug(logMessage, "receive", true)
	}

	err = json.Unmarshal(out.Body, &response)
	if err != nil {
		return nil, errors.E(op, err)
	}

	for _, f := range response {
		msg, err := c.parseFrame(f)
		if err != nil {
			return nil, errors.E(op, err)
		}

		result = append(result, msg)
	}

	return result, nil
}

func (c *JSONCodec) packFrame(msg Message) (jsonFrame, error) {
	var (
		err   error
		frame jsonFrame
	)

	frame.ID = msg.ID

	if msg.Payloads != nil {
		frame.Payloads, err = msg.Payloads.Marshal()
		if err != nil {
			return jsonFrame{}, err
		}
	}

	if msg.Failure != nil {
		frame.Failure, err = msg.Failure.Marshal()
		if err != nil {
			return jsonFrame{}, err
		}
	}

	if msg.Command == nil {
		return frame, nil
	}

	frame.Command, err = commandName(msg.Command)
	if err != nil {
		return jsonFrame{}, err
	}

	frame.Options, err = json.Marshal(msg.Command)
	if err != nil {
		return jsonFrame{}, err
	}

	return frame, nil
}

func (c *JSONCodec) parseFrame(frame jsonFrame) (Message, error) {
	var (
		err error
		msg Message
	)

	msg.ID = frame.ID

	if frame.Payloads != nil {
		msg.Payloads = &common.Payloads{}

		err = msg.Payloads.Unmarshal(frame.Payloads)
		if err != nil {
			return Message{}, err
		}
	}

	if frame.Failure != nil {
		msg.Failure = &failure.Failure{}

		err = msg.Failure.Unmarshal(frame.Failure)
		if err != nil {
			return Message{}, err
		}
	}

	if frame.Command != "" {
		cmd, err := initCommand(frame.Command)
		if err != nil {
			return Message{}, err
		}

		err = json.Unmarshal(frame.Options, &cmd)
		if err != nil {
			return Message{}, err
		}

		msg.Command = cmd
	}

	return msg, nil
}
