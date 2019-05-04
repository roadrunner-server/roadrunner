package limit

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
	cfg := &mockCfg{`{"enable: true}`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}
