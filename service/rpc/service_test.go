package rpc

import (
	"encoding/json"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type testService struct{}

func (ts *testService) Echo(msg string, r *string) error { *r = msg; return nil }

type testCfg struct{ cfg string }

func (cfg *testCfg) Get(name string) service.Config  { return nil }
func (cfg *testCfg) Unmarshal(out interface{}) error { return json.Unmarshal([]byte(cfg.cfg), out) }

func Test_ConfigError(t *testing.T) {
	s := &Service{}
	ok, err := s.Init(&testCfg{`{"enable":false`}, nil)

	assert.Error(t, err)
	assert.False(t, ok)
}

func Test_Disabled(t *testing.T) {
	s := &Service{}
	ok, err := s.Init(&testCfg{`{"enable":false}`}, nil)

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
	ok, err := s.Init(&testCfg{`{"enable":true, "listen":"tcp://localhost:9008"}`}, nil)

	assert.NoError(t, err)
	assert.True(t, ok)
}

func Test_StopNonServing(t *testing.T) {
	s := &Service{}
	ok, err := s.Init(&testCfg{`{"enable":true, "listen":"tcp://localhost:9008"}`}, nil)

	assert.NoError(t, err)
	assert.True(t, ok)
	s.Stop()
}

func Test_Serve_Errors(t *testing.T) {
	s := &Service{}
	ok, err := s.Init(&testCfg{`{"enable":true, "listen":"mailformed"}`}, nil)
	assert.NoError(t, err)
	assert.True(t, ok)

	assert.Error(t, s.Serve())

	client, err := s.Client()
	assert.Nil(t, client)
	assert.Error(t, err)
}

func Test_Serve_Client(t *testing.T) {
	s := &Service{}
	ok, err := s.Init(&testCfg{`{"enable":true, "listen":"tcp://localhost:9018"}`}, nil)
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
