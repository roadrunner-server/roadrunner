package headers

import (
	"encoding/json"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockCfg struct{ cfg string }

func (cfg *mockCfg) Get(name string) service.Config  { return nil }
func (cfg *mockCfg) Unmarshal(out interface{}) error { return json.Unmarshal([]byte(cfg.cfg), out) }

func Test_Config_Hydrate_Error1(t *testing.T) {
	cfg := &mockCfg{`{"request": {"From": "Something"}}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))
}

func Test_Config_Hydrate_Error2(t *testing.T) {
	cfg := &mockCfg{`{"dir": "/dir/"`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}
