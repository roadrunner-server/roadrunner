package sqs

import (
	json "github.com/json-iterator/go"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockCfg struct{ cfg string }

func (cfg *mockCfg) Get(name string) service.Config  { return nil }
func (cfg *mockCfg) Unmarshal(out interface{}) error { return json.Unmarshal([]byte(cfg.cfg), out) }

func Test_Config_Hydrate_Error(t *testing.T) {
	cfg := &mockCfg{`{"dead`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func Test_Config_Hydrate_Error2(t *testing.T) {
	cfg := &mockCfg{`{}`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func Test_Config_Hydrate_Error3(t *testing.T) {
	cfg := &mockCfg{`{"region":"us-east-1"}`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func Test_Config_Hydrate_Error4(t *testing.T) {
	cfg := &mockCfg{`{"region":"us-east-1","key":"key"}`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func Test_Config_Hydrate_Error5(t *testing.T) {
	cfg := &mockCfg{`{"region":"us-east-1","key":"key","secret":"secret"}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))
}
