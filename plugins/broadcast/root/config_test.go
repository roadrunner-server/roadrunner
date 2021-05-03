package broadcast

import (
	"encoding/json"
	"testing"

	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/rpc"
	"github.com/stretchr/testify/assert"
)

type testCfg struct {
	rpc       string
	broadcast string
	target    string
}

func (cfg *testCfg) Get(name string) service.Config {
	if name == ID {
		return &testCfg{target: cfg.broadcast}
	}

	if name == rpc.ID {
		return &testCfg{target: cfg.rpc}
	}

	return nil
}

func (cfg *testCfg) Unmarshal(out interface{}) error {
	return json.Unmarshal([]byte(cfg.target), out)
}

func Test_Config_Hydrate_Error(t *testing.T) {
	cfg := &testCfg{target: `{"dead`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func Test_Config_Hydrate_OK(t *testing.T) {
	cfg := &testCfg{target: `{"path":"/path"}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))
}

func Test_Config_Redis_Error(t *testing.T) {
	cfg := &testCfg{target: `{"path":"/path","redis":{}}`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func Test_Config_Redis_OK(t *testing.T) {
	cfg := &testCfg{target: `{"path":"/path","redis":{"addr":"localhost:6379"}}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))
}
