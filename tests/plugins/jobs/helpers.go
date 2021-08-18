package jobs

import (
	"bytes"
	"net"
	"net/http"
	"net/rpc"
	"testing"

	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	jobState "github.com/spiral/roadrunner/v2/pkg/state/job"
	jobsv1beta "github.com/spiral/roadrunner/v2/proto/jobs/v1beta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	push    string = "jobs.Push"
	pause   string = "jobs.Pause"
	destroy string = "jobs.Destroy"
	resume  string = "jobs.Resume"
	stat    string = "jobs.Stat"
)

func resumePipes(pipes ...string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		assert.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		pipe := &jobsv1beta.Pipelines{Pipelines: make([]string, len(pipes))}

		for i := 0; i < len(pipes); i++ {
			pipe.GetPipelines()[i] = pipes[i]
		}

		er := &jobsv1beta.Empty{}
		err = client.Call(resume, pipe, er)
		assert.NoError(t, err)
	}
}

func pushToDisabledPipe(pipeline string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		assert.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		req := &jobsv1beta.PushRequest{Job: &jobsv1beta.Job{
			Job:     "some/php/namespace",
			Id:      "1",
			Payload: `{"hello":"world"}`,
			Headers: nil,
			Options: &jobsv1beta.Options{
				Priority: 1,
				Pipeline: pipeline,
			},
		}}

		er := &jobsv1beta.Empty{}
		err = client.Call(push, req, er)
		assert.NoError(t, err)
	}
}

func pushToPipe(pipeline string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		assert.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		req := &jobsv1beta.PushRequest{Job: &jobsv1beta.Job{
			Job:     "some/php/namespace",
			Id:      "1",
			Payload: `{"hello":"world"}`,
			Headers: map[string]*jobsv1beta.HeaderValue{"test": {Value: []string{"test2"}}},
			Options: &jobsv1beta.Options{
				Priority: 1,
				Pipeline: pipeline,
				Delay:    0,
			},
		}}

		er := &jobsv1beta.Empty{}
		err = client.Call(push, req, er)
		assert.NoError(t, err)
	}
}

func pushToPipeDelayed(pipeline string, delay int64) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		assert.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		req := &jobsv1beta.PushRequest{Job: &jobsv1beta.Job{
			Job:     "some/php/namespace",
			Id:      "1",
			Payload: `{"hello":"world"}`,
			Headers: map[string]*jobsv1beta.HeaderValue{"test": {Value: []string{"test2"}}},
			Options: &jobsv1beta.Options{
				Priority: 1,
				Pipeline: pipeline,
				Delay:    delay,
			},
		}}

		er := &jobsv1beta.Empty{}
		err = client.Call(push, req, er)
		assert.NoError(t, err)
	}
}

func pushToPipeErr(pipeline string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		require.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		req := &jobsv1beta.PushRequest{Job: &jobsv1beta.Job{
			Job:     "some/php/namespace",
			Id:      "1",
			Payload: `{"hello":"world"}`,
			Headers: map[string]*jobsv1beta.HeaderValue{"test": {Value: []string{"test2"}}},
			Options: &jobsv1beta.Options{
				Priority: 1,
				Pipeline: pipeline,
				Delay:    0,
			},
		}}

		er := &jobsv1beta.Empty{}
		err = client.Call(push, req, er)
		require.Error(t, err)
	}
}
func pausePipelines(pipes ...string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		assert.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		pipe := &jobsv1beta.Pipelines{Pipelines: make([]string, len(pipes))}

		for i := 0; i < len(pipes); i++ {
			pipe.GetPipelines()[i] = pipes[i]
		}

		er := &jobsv1beta.Empty{}
		err = client.Call(pause, pipe, er)
		assert.NoError(t, err)
	}
}

func destroyPipelines(pipes ...string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		assert.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		pipe := &jobsv1beta.Pipelines{Pipelines: make([]string, len(pipes))}

		for i := 0; i < len(pipes); i++ {
			pipe.GetPipelines()[i] = pipes[i]
		}

		er := &jobsv1beta.Empty{}
		err = client.Call(destroy, pipe, er)
		assert.NoError(t, err)
	}
}

func enableProxy(name string, t *testing.T) {
	buf := new(bytes.Buffer)
	buf.WriteString(`{"enabled":true}`)

	resp, err := http.Post("http://127.0.0.1:8474/proxies/"+name, "application/json", buf) //nolint:noctx
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	if resp.Body != nil {
		_ = resp.Body.Close()
	}
}

func disableProxy(name string, t *testing.T) {
	buf := new(bytes.Buffer)
	buf.WriteString(`{"enabled":false}`)

	resp, err := http.Post("http://127.0.0.1:8474/proxies/"+name, "application/json", buf) //nolint:noctx
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	if resp.Body != nil {
		_ = resp.Body.Close()
	}
}

func deleteProxy(name string, t *testing.T) {
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodDelete, "http://127.0.0.1:8474/proxies/"+name, nil) //nolint:noctx
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, 204, resp.StatusCode)
	if resp.Body != nil {
		_ = resp.Body.Close()
	}
}

func stats(state *jobState.State) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		assert.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		st := &jobsv1beta.Stats{}
		er := &jobsv1beta.Empty{}

		err = client.Call(stat, er, st)
		require.NoError(t, err)
		require.NotNil(t, st)

		state.Queue = st.Stats[0].Queue
		state.Pipeline = st.Stats[0].Pipeline
		state.Driver = st.Stats[0].Driver
		state.Active = st.Stats[0].Active
		state.Delayed = st.Stats[0].Delayed
		state.Reserved = st.Stats[0].Reserved
		state.Ready = st.Stats[0].Ready
	}
}
