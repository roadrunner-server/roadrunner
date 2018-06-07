package rpc

import (
	"encoding/json"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testService struct{}

func (ts *testService) Echo(msg string, r *string) error { *r = msg; return nil }

type testCfg struct{ cfg string }

func (cfg *testCfg) Get(name string) service.Config  { return nil }
func (cfg *testCfg) Unmarshal(out interface{}) error { return json.Unmarshal([]byte(cfg.cfg), out) }

func Test_Disabled(t *testing.T) {
	s, err := (&Service{}).WithConfig(&testCfg{`{"enable":false}`}, nil)

	assert.NoError(t, err)
	assert.Nil(t, s)
}

func Test_RegisterNotConfigured(t *testing.T) {
	s := &Service{}
	assert.Error(t, s.Register("test", &testService{}))

	client, err := s.Client()
	assert.Nil(t, client)
	assert.Error(t, err)
}

func Test_Enabled(t *testing.T) {
	s, err := (&Service{}).WithConfig(&testCfg{`{"enable":true, "listen":"tcp://localhost:9008"}`}, nil)

	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.IsType(t, &Service{}, s)
}

func Test_StopNonServing(t *testing.T) {
	s, err := (&Service{}).WithConfig(&testCfg{`{"enable":true, "listen":"tcp://localhost:9008"}`}, nil)

	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.IsType(t, &Service{}, s)
	s.Stop()
}

func Test_Serve_Errors(t *testing.T) {
	s, err := (&Service{}).WithConfig(&testCfg{`{"enable":true, "listen":"mailformed"}`}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.IsType(t, &Service{}, s)

	assert.Error(t, s.Serve())

	client, err := s.(*Service).Client()
	assert.Nil(t, client)
	assert.Error(t, err)
}

func Test_Serve_Client(t *testing.T) {
	s, err := (&Service{}).WithConfig(&testCfg{`{"enable":true, "listen":"tcp://localhost:9008"}`}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.IsType(t, &Service{}, s)
	defer s.Stop()

	assert.NoError(t, s.(*Service).Register("test", &testService{}))

	go func() { assert.NoError(t, s.Serve()) }()

	client, err := s.(*Service).Client()
	assert.NotNil(t, client)
	assert.NoError(t, err)
	defer client.Close()

	var resp string
	assert.NoError(t, client.Call("test.Echo", "hello world", &resp))
	assert.Equal(t, "hello world", resp)
}
