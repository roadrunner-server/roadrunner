package env

import (
	"encoding/json"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockCfg struct{ cfg string }

func (cfg *mockCfg) Get(name string) service.Config  { return nil }
func (cfg *mockCfg) Unmarshal(out interface{}) error { return json.Unmarshal([]byte(cfg.cfg), out) }

func Test_Config_Hydrate(t *testing.T) {
	cfg := &mockCfg{`{"key":"value"}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))
	assert.Len(t, c.Values, 1)
}

func Test_Config_Hydrate_Empty(t *testing.T) {
	cfg := &mockCfg{`{}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))
	assert.Len(t, c.Values, 0)
}

func Test_Config_Defaults(t *testing.T) {
	c := &Config{}
	c.InitDefaults()
	assert.Len(t, c.Values, 0)
}
