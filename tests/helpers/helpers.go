package helpers

import (
	"context"
	"net"
	"net/rpc"
	"testing"
	"time"

	"github.com/google/uuid"
	jobsProto "github.com/roadrunner-server/api/v4/build/jobs/v1"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	"github.com/stretchr/testify/require"
)

const (
	push    = "jobs.Push"
	pause   = "jobs.Pause"
	destroy = "jobs.Destroy"
	resume  = "jobs.Resume"

	dialTimeout = 5 * time.Second
)

// rpcClient dials the given address with a timeout and returns a Goridge RPC
// client. The client is automatically closed via t.Cleanup when the test ends.
func rpcClient(t *testing.T, address string) *rpc.Client {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	conn, err := new(net.Dialer).DialContext(ctx, "tcp", address)
	require.NoError(t, err)

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	t.Cleanup(func() { _ = client.Close() })

	return client
}

// callPipelines is a generic helper that calls the given RPC method
// on the specified pipelines.
func callPipelines(t *testing.T, address, method string, pipes []string) {
	t.Helper()

	client := rpcClient(t, address)

	pipe := &jobsProto.Pipelines{Pipelines: make([]string, len(pipes))}
	for i := range pipes {
		pipe.GetPipelines()[i] = pipes[i]
	}

	er := &jobsProto.Empty{}
	err := client.Call(method, pipe, er)
	require.NoError(t, err)
}

// ResumePipes resumes the specified pipelines via RPC.
func ResumePipes(address string, pipes ...string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		callPipelines(t, address, resume, pipes)
	}
}

// PausePipelines pauses the specified pipelines via RPC.
func PausePipelines(address string, pipes ...string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		callPipelines(t, address, pause, pipes)
	}
}

// PushToPipe pushes a single job to the specified pipeline via RPC.
func PushToPipe(pipeline string, autoAck bool, address string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		client := rpcClient(t, address)

		req := &jobsProto.PushRequest{Job: &jobsProto.Job{
			Job:     "some/php/namespace",
			Id:      uuid.NewString(),
			Payload: []byte(`{"hello":"world"}`),
			Headers: map[string]*jobsProto.HeaderValue{"test": {Value: []string{"test2"}}},
			Options: &jobsProto.Options{
				AutoAck:  autoAck,
				Priority: 1,
				Pipeline: pipeline,
				Topic:    pipeline,
			},
		}}

		er := &jobsProto.Empty{}
		err := client.Call(push, req, er)
		require.NoError(t, err)
	}
}

// DestroyPipelines destroys the specified pipelines via RPC, retrying up to 10 times.
func DestroyPipelines(address string, pipes ...string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		client := rpcClient(t, address)

		pipe := &jobsProto.Pipelines{Pipelines: make([]string, len(pipes))}
		for i := range pipes {
			pipe.GetPipelines()[i] = pipes[i]
		}

		var lastErr error
		for range 10 {
			er := &jobsProto.Empty{}
			lastErr = client.Call(destroy, pipe, er)
			if lastErr != nil {
				time.Sleep(time.Second)
				continue
			}
			return
		}
		require.NoError(t, lastErr)
	}
}
