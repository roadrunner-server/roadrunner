package rpc

import (
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/env"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type testService struct{}

func (ts *testService) Echo(msg string, r *string) error { *r = msg; return nil }

func Test_Disabled(t *testing.T) {
	s := &Service{}
	ok, err := s.Init(&Config{Enable: false}, service.NewContainer(nil), nil)

	assert.NoError(t, err)
	assert.False(t, ok)
}

func Test_RegisterNotConfigured(t *testing.T) {
	s := &Service{}
	assert.Error(t, s.Register("test", &testService{}))

	client, err := s.Client()
	assert.Nil(t, client)
	assert.Error(t, err)
	assert.Error(t, s.Serve())
}

func Test_Enabled(t *testing.T) {
	s := &Service{}
	ok, err := s.Init(&Config{Enable: true, Listen: "tcp://localhost:9008"}, service.NewContainer(nil), nil)

	assert.NoError(t, err)
	assert.True(t, ok)
}

func Test_StopNonServing(t *testing.T) {
	s := &Service{}
	ok, err := s.Init(&Config{Enable: true, Listen: "tcp://localhost:9008"}, service.NewContainer(nil), nil)

	assert.NoError(t, err)
	assert.True(t, ok)
	s.Stop()
}

func Test_Serve_Errors(t *testing.T) {
	s := &Service{}
	ok, err := s.Init(&Config{Enable: true, Listen: "mailformed"}, service.NewContainer(nil), nil)
	assert.NoError(t, err)
	assert.True(t, ok)

	assert.Error(t, s.Serve())

	client, err := s.Client()
	assert.Nil(t, client)
	assert.Error(t, err)
}

func Test_Serve_Client(t *testing.T) {
	s := &Service{}
	ok, err := s.Init(&Config{Enable: true, Listen: "tcp://localhost:9018"}, service.NewContainer(nil), nil)
	assert.NoError(t, err)
	assert.True(t, ok)

	defer s.Stop()

	assert.NoError(t, s.Register("test", &testService{}))

	go func() { assert.NoError(t, s.Serve()) }()
	time.Sleep(time.Millisecond)

	client, err := s.Client()
	assert.NotNil(t, client)
	assert.NoError(t, err)
	defer client.Close()

	var resp string
	assert.NoError(t, client.Call("test.Echo", "hello world", &resp))
	assert.Equal(t, "hello world", resp)
}

func TestSetEnv(t *testing.T) {
	s := &Service{}
	e := env.NewService(map[string]string{})
	ok, err := s.Init(&Config{Enable: true, Listen: "tcp://localhost:9018"}, service.NewContainer(nil), e)

	assert.NoError(t, err)
	assert.True(t, ok)

	v, _ := e.GetEnv()
	assert.Equal(t, "tcp://localhost:9018", v["RR_RPC"])
}
