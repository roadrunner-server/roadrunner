package metrics

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
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

func Test_Config_Metrics(t *testing.T) {
	cfg := &mockCfg{`{
"collect":{
	"metric1":{"type": "gauge"},
	"metric2":{	"type": "counter"},
	"metric3":{"type": "summary"},
	"metric4":{"type": "histogram"}
}
}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))

	m, err := c.getCollectors()
	assert.NoError(t, err)

	assert.IsType(t, prometheus.NewGauge(prometheus.GaugeOpts{}), m["metric1"])
	assert.IsType(t, prometheus.NewCounter(prometheus.CounterOpts{}), m["metric2"])
	assert.IsType(t, prometheus.NewSummary(prometheus.SummaryOpts{}), m["metric3"])
	assert.IsType(t, prometheus.NewHistogram(prometheus.HistogramOpts{}), m["metric4"])
}

func Test_Config_MetricsVector(t *testing.T) {
	cfg := &mockCfg{`{
"collect":{
	"metric1":{"type": "gauge","labels":["label"]},
	"metric2":{	"type": "counter","labels":["label"]},
	"metric3":{"type": "summary","labels":["label"]},
	"metric4":{"type": "histogram","labels":["label"]}
}
}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))

	m, err := c.getCollectors()
	assert.NoError(t, err)

	assert.IsType(t, prometheus.NewGaugeVec(prometheus.GaugeOpts{}, []string{}), m["metric1"])
	assert.IsType(t, prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{}), m["metric2"])
	assert.IsType(t, prometheus.NewSummaryVec(prometheus.SummaryOpts{}, []string{}), m["metric3"])
	assert.IsType(t, prometheus.NewHistogramVec(prometheus.HistogramOpts{}, []string{}), m["metric4"])
}
