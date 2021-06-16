package oooold

import (
	json "github.com/json-iterator/go"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockCfg struct{ cfg string }

func (cfg *mockCfg) Get(name string) service.Config {
	if name == "same" || name == "jobs" {
		return cfg
	}

	return nil
}
func (cfg *mockCfg) Unmarshal(out interface{}) error { return json.Unmarshal([]byte(cfg.cfg), out) }

func Test_Config_Hydrate_Error(t *testing.T) {
	cfg := &mockCfg{cfg: `{"dead`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func Test_Config_Hydrate_OK(t *testing.T) {
	cfg := &mockCfg{cfg: `{
	"workers":{"pool":{"numWorkers": 1}}
}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))
}

func Test_Config_Hydrate_Unmarshal(t *testing.T) {
	cfg := &mockCfg{cfg: `{
	"workers":{"pool":{"numWorkers": 1}}
}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))

	var i interface{}
	assert.Nil(t, c.Unmarshal(i))
}

func Test_Config_Hydrate_Get(t *testing.T) {
	cfg := &mockCfg{cfg: `{
	"workers":{"pool":{"numWorkers": 1}}
}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))

	assert.Nil(t, c.Get("nil"))
}

func Test_Config_Hydrate_Get_Valid(t *testing.T) {
	cfg := &mockCfg{cfg: `{
	"workers":{"pool":{"numWorkers": 1}}
}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))

	assert.Equal(t, cfg, c.Get("same"))
}

func Test_Config_Hydrate_GetNoParent(t *testing.T) {
	c := &Config{}
	assert.Nil(t, c.Get("nil"))
}

func Test_Pipelines(t *testing.T) {
	cfg := &mockCfg{cfg: `{
	"workers":{
		"pool":{"numWorkers": 1}
	},
	"pipelines":{
		"pipe": {"broker":"broker"}
	},
	"dispatch":{
		"job.*": {"pipeline":"default"}
	}
	}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))

	assert.Equal(t, "pipe", c.pipelines.Get("pipe").Name())
	assert.Equal(t, "broker", c.pipelines.Get("pipe").Broker())
}

func Test_Pipelines_NoBroker(t *testing.T) {
	cfg := &mockCfg{cfg: `{
	"workers":{
		"pool":{"numWorkers": 1}
	},
	"pipelines":{
		"pipe": {}
	},
	"dispatch":{
		"job.*": {"pipeline":"default"}
	}
	}`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func Test_MatchPipeline(t *testing.T) {
	cfg := &mockCfg{cfg: `{
	"workers":{
		"pool":{"numWorkers": 1}
	},
	"pipelines":{
		"pipe": {"broker":"default"}
	},
	"dispatch":{
		"job.*": {"pipeline":"pipe","delay":10}
	}
	}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))

	_, _, err := c.MatchPipeline(&Job{Job: "undefined", Options: &Options{}})
	assert.Error(t, err)

	p, _, _ := c.MatchPipeline(&Job{Job: "undefined", Options: &Options{Pipeline: "pipe"}})
	assert.Equal(t, "pipe", p.Name())

	p, opt, _ := c.MatchPipeline(&Job{Job: "job.abc", Options: &Options{}})
	assert.Equal(t, "pipe", p.Name())
	assert.Equal(t, 10, opt.Delay)
}

func Test_MatchPipeline_Error(t *testing.T) {
	cfg := &mockCfg{cfg: `{
	"workers":{
		"pool":{"numWorkers": 1}
	},
	"pipelines":{
		"pipe": {"broker":"default"}
	},
	"dispatch":{
		"job.*": {"pipeline":"missing"}
	}
	}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))

	_, _, err := c.MatchPipeline(&Job{Job: "job.abc", Options: &Options{}})
	assert.Error(t, err)
}
