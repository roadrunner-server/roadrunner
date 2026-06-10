package helpers

import (
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	jobsV2 "github.com/roadrunner-server/api-go/v6/jobs/v2"
	"github.com/roadrunner-server/api-go/v6/jobs/v2/jobsV2connect"
	"github.com/stretchr/testify/require"
)

// jobsClient returns a JobsService connect client bound to the RoadRunner
// Connect-RPC plane listening on the given tcp address.
func jobsClient(t *testing.T, address string) jobsV2connect.JobsServiceClient {
	t.Helper()

	return jobsV2connect.NewJobsServiceClient(http.DefaultClient, "http://"+address)
}

// ResumePipes resumes the specified pipelines via RPC.
func ResumePipes(address string, pipes ...string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		_, err := jobsClient(t, address).Resume(t.Context(), connect.NewRequest(&jobsV2.Pipelines{Pipelines: pipes}))
		require.NoError(t, err)
	}
}

// PausePipelines pauses the specified pipelines via RPC.
func PausePipelines(address string, pipes ...string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		_, err := jobsClient(t, address).Pause(t.Context(), connect.NewRequest(&jobsV2.Pipelines{Pipelines: pipes}))
		require.NoError(t, err)
	}
}

// PushToPipe pushes a single job to the specified pipeline via RPC.
func PushToPipe(pipeline string, autoAck bool, address string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		req := &jobsV2.PushRequest{Job: &jobsV2.Job{
			Job:     "some/php/namespace",
			Id:      uuid.NewString(),
			Payload: []byte(`{"hello":"world"}`),
			Headers: map[string]*jobsV2.JobHeaderValue{"test": {Values: []string{"test2"}}},
			Options: &jobsV2.Options{
				AutoAck:  autoAck,
				Priority: 1,
				Pipeline: pipeline,
				Topic:    pipeline,
			},
		}}

		_, err := jobsClient(t, address).Push(t.Context(), connect.NewRequest(req))
		require.NoError(t, err)
	}
}

// DestroyPipelines destroys the specified pipelines via RPC, retrying up to 10 times.
func DestroyPipelines(address string, pipes ...string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		client := jobsClient(t, address)
		pipe := &jobsV2.Pipelines{Pipelines: pipes}

		var lastErr error
		for range 10 {
			_, lastErr = client.Destroy(t.Context(), connect.NewRequest(pipe))
			if lastErr != nil {
				time.Sleep(time.Second)
				continue
			}
			return
		}
		require.NoError(t, lastErr)
	}
}
