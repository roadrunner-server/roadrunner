package limit

import (
	"testing"
	"time"

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
	cfg := &mockCfg{`{"enable: true}`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func Test_Controller_Default(t *testing.T) {
	cfg := &mockCfg{`
{
	"services":{
		"http": {
			"TTL": 1
		}
	}
}
`}
	c := &Config{}
	err := c.InitDefaults()
	if err != nil {
		t.Errorf("failed to InitDefaults: error %v", err)
	}

	assert.NoError(t, c.Hydrate(cfg))
	assert.Equal(t, time.Second, c.Interval)

	list := c.Controllers(func(event int, ctx interface{}) {
	})

	sc := list["http"]

	assert.Equal(t, time.Second, sc.(*controller).tick)
}
