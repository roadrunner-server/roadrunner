package tests

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tests/helpers"
	mocklogger "tests/mock"

	pb "github.com/roadrunner-server/api-go/v6/jobs/v1"
	"github.com/roadrunner-server/config/v6"
	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/roadrunner/v3/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisJobsPHPRequeue(t *testing.T) {
	queue := rand.Text()
	result := filepath.Join(t.TempDir(), "receipt.json")
	t.Setenv("RR_REDIS_JOB_RESULT", result)
	t.Setenv("RR_REDIS_JOB_QUEUE", queue)
	cfg := &config.Plugin{Path: "configs/.rr-jobs-redis.yaml"}
	logger, _ := mocklogger.SlogTestLogger(slog.LevelDebug)
	cont := endure.New(slog.LevelError)
	require.NoError(t, cont.RegisterAll(append(container.Plugins(), cfg, logger)...))
	require.NoError(t, cont.Init())
	_, err := cont.Serve()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cont.Stop()) })

	client := helpers.RPCClient(t, "127.0.0.1:6209")
	req := &pb.PushRequest{Job: &pb.Job{Id: "123", Job: "send", Payload: []byte{0, 255, 42},
		Headers: map[string]*pb.HeaderValue{"tags": {Value: []string{"one", "two"}}}, Options: &pb.Options{Pipeline: "mail"}}}
	require.NoError(t, client.Call("jobs.Push", req, &pb.Empty{}))
	stats := &pb.Stats{}
	require.NoError(t, client.Call("jobs.Stat", &pb.Empty{}, stats))
	require.Len(t, stats.GetStats(), 1)
	require.False(t, stats.GetStats()[0].GetReady())
	require.EqualValues(t, 1, stats.GetStats()[0].GetReserved())

	require.NoError(t, client.Call("jobs.Resume", &pb.Pipelines{Pipelines: []string{"mail"}}, &pb.Pipelines{}))
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		out, err := os.ReadFile(result)
		require.NoError(c, err)
		require.JSONEq(c, fmt.Sprintf(`["123","redis","mail",%q,"AP8q",["one","two"],["yes"]]`, queue), string(out))
		state := &pb.Stats{}
		require.NoError(c, client.Call("jobs.Stat", &pb.Empty{}, state))
		require.Len(c, state.GetStats(), 1)
		s := state.GetStats()[0]
		require.Zero(c, s.GetActive()+s.GetReserved()+s.GetDelayed())
	}, 10*time.Second, 20*time.Millisecond)
}
