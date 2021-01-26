package protocol

import (
	"github.com/spiral/roadrunner/v2/pkg/payload"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/failure/v1"
)

const (
	// DebugNone disables all debug messages.
	DebugNone = iota

	// DebugNormal renders all messages into console.
	DebugNormal

	// DebugHumanized enables color highlights for messages.
	DebugHumanized
)

// Context provides worker information about currently. Context can be empty for server level commands.
type Context struct {
	// TaskQueue associates message batch with the specific task queue in underlying worker.
	TaskQueue string `json:"taskQueue,omitempty"`

	// TickTime associated current or historical time with message batch.
	TickTime string `json:"tickTime,omitempty"`

	// Replay indicates that current message batch is historical.
	Replay bool `json:"replay,omitempty"`
}

// Message used to exchange the send commands and receive responses from underlying workers.
type Message struct {
	// ID contains ID of the command, response or error.
	ID uint64 `json:"id"`

	// Command of the message in unmarshalled form. Pointer.
	Command interface{} `json:"command,omitempty"`

	// Failure associated with command id.
	Failure *failure.Failure `json:"failure,omitempty"`

	// Payloads contains message specific payloads in binary format.
	Payloads *commonpb.Payloads `json:"payloads,omitempty"`
}

// Codec manages payload encoding and decoding while communication with underlying worker.
type Codec interface {
	// WithLogger creates new codes instance with attached logger.
	WithLogger(logger.Logger) Codec

	// GetName returns codec name.
	GetName() string

	// Execute sends message to worker and waits for the response.
	Execute(e Endpoint, ctx Context, msg ...Message) ([]Message, error)
}

// Endpoint provides the ability to send and receive messages.
type Endpoint interface {
	// ExecWithContext allow to set ExecTTL
	Exec(p payload.Payload) (payload.Payload, error)
}

// DebugLevel configures debug level.
type DebugLevel int

// IsEmpty only check if task queue set.
func (ctx Context) IsEmpty() bool {
	return ctx.TaskQueue == ""
}

// IsCommand returns true if message carries request.
func (msg Message) IsCommand() bool {
	return msg.Command != nil
}
