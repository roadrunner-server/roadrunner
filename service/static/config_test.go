package static

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
	cfg := &mockCfg{`{"dir": "./"}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))
}

func Test_Config_Hydrate_Error(t *testing.T) {
	cfg := &mockCfg{`{"enable": true,"dir": "/dir/"}`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func TestConfig_Forbids(t *testing.T) {
	cfg := Config{Forbid: []string{".php"}}

	assert.True(t, cfg.AlwaysForbid("index.php"))
	assert.True(t, cfg.AlwaysForbid("index.PHP"))
	assert.True(t, cfg.AlwaysForbid("phpadmin/index.bak.php"))
	assert.False(t, cfg.AlwaysForbid("index.html"))
}

func TestConfig_Valid(t *testing.T) {
	assert.NoError(t, (&Config{Dir: "./"}).Valid())
	assert.Error(t, (&Config{Dir: "./config.go"}).Valid())
	assert.Error(t, (&Config{Dir: "./dir/"}).Valid())
}
