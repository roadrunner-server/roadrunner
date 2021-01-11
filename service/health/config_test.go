package health

import (
	"testing"

	json "github.com/json-iterator/go"

	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
)

type mockCfg struct{ cfg string }

func (cfg *mockCfg) Get(name string) service.Config { return nil }
func (cfg *mockCfg) Unmarshal(out interface{}) error {
	j := json.ConfigCompatibleWithStandardLibrary
	return j.Unmarshal([]byte(cfg.cfg), out)
}

func Test_Config_Hydrate_Error1(t *testing.T) {
	cfg := &mockCfg{`{"address": "localhost:8080"}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))
	assert.Equal(t, "localhost:8080", c.Address)
}

func Test_Config_Hydrate_Error2(t *testing.T) {
	cfg := &mockCfg{`{"dir": "/dir/"`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func Test_Config_Hydrate_Valid1(t *testing.T) {
	cfg := &mockCfg{`{"address": "localhost"}`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func Test_Config_Hydrate_Valid2(t *testing.T) {
	cfg := &mockCfg{`{"address": ":1111"}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))
}
