package activity

import (
	v1Proto "github.com/golang/protobuf/proto" //nolint:staticcheck
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/proto"
)

/*
- the method's type is exported.
- the method is exported.
- the method has two arguments, both exported (or builtin) types.
- the method's second argument is a pointer.
- the method has return type error.
*/
type rpc struct {
	srv    *Plugin
	client client.Client
}

// RecordHeartbeatRequest sent by activity to record current state.
type RecordHeartbeatRequest struct {
	TaskToken []byte `json:"taskToken"`
	Details   []byte `json:"details"`
}

// RecordHeartbeatResponse sent back to the worker to indicate that activity was cancelled.
type RecordHeartbeatResponse struct {
	Canceled bool `json:"canceled"`
}

// RecordActivityHeartbeat records heartbeat for an activity.
// taskToken - is the value of the binary "TaskToken" field of the "ActivityInfo" struct retrieved inside the activity.
// details - is the progress you want to record along with heart beat for this activity.
// The errors it can return:
// - EntityNotExistsError
// - InternalServiceError
// - CanceledError
func (r *rpc) RecordActivityHeartbeat(in RecordHeartbeatRequest, out *RecordHeartbeatResponse) error {
	details := &commonpb.Payloads{}

	if len(in.Details) != 0 {
		if err := proto.Unmarshal(in.Details, v1Proto.MessageV2(details)); err != nil {
			return err
		}
	}

	// find running activity
	ctx, err := r.srv.getPool().GetActivityContext(in.TaskToken)
	if err != nil {
		return err
	}

	activity.RecordHeartbeat(ctx, details)

	select {
	case <-ctx.Done():
		*out = RecordHeartbeatResponse{Canceled: true}
	default:
		*out = RecordHeartbeatResponse{Canceled: false}
	}

	return nil
}
