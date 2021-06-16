package beanstalk

import (
	json "github.com/json-iterator/go"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockCfg struct{ cfg string }

func (cfg *mockCfg) Get(name string) service.Config  { return nil }
func (cfg *mockCfg) Unmarshal(out interface{}) error { return json.Unmarshal([]byte(cfg.cfg), out) }

func TestConfig_Hydrate_Error(t *testing.T) {
	cfg := &mockCfg{`{"dead`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func TestConfig_Hydrate_Error2(t *testing.T) {
	cfg := &mockCfg{`{"addr":""}`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func TestConfig_Hydrate_Error3(t *testing.T) {
	cfg := &mockCfg{`{"addr":"tcp"}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))

	_, err := c.newConn()
	assert.Error(t, err)
}

func TestConfig_Hydrate_Error4(t *testing.T) {
	cfg := &mockCfg{`{"addr":"unix://sock.bean"}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))

	_, err := c.newConn()
	assert.Error(t, err)
}
