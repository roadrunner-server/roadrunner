package metrics

import (
	"bytes"
	"testing"

	j "github.com/json-iterator/go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

var json = j.ConfigCompatibleWithStandardLibrary

func Test_Config_Hydrate_Error1(t *testing.T) {
	cfg := `{"request": {"From": "Something"}}`
	c := &Config{}
	f := new(bytes.Buffer)
	f.WriteString(cfg)

	err := json.Unmarshal(f.Bytes(), &c)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Config_Hydrate_Error2(t *testing.T) {
	cfg := `{"dir": "/dir/"`
	c := &Config{}

	f := new(bytes.Buffer)
	f.WriteString(cfg)

	err := json.Unmarshal(f.Bytes(), &c)
	assert.Error(t, err)
}

func Test_Config_Metrics(t *testing.T) {
	cfg := `{
"collect":{
	"metric1":{"type": "gauge"},
	"metric2":{	"type": "counter"},
	"metric3":{"type": "summary"},
	"metric4":{"type": "histogram"}
}
}`
	c := &Config{}
	f := new(bytes.Buffer)
	f.WriteString(cfg)

	err := json.Unmarshal(f.Bytes(), &c)
	if err != nil {
		t.Fatal(err)
	}

	m, err := c.getCollectors()
	assert.NoError(t, err)

	assert.IsType(t, prometheus.NewGauge(prometheus.GaugeOpts{}), m["metric1"])
	assert.IsType(t, prometheus.NewCounter(prometheus.CounterOpts{}), m["metric2"])
	assert.IsType(t, prometheus.NewSummary(prometheus.SummaryOpts{}), m["metric3"])
	assert.IsType(t, prometheus.NewHistogram(prometheus.HistogramOpts{}), m["metric4"])
}

func Test_Config_MetricsVector(t *testing.T) {
	cfg := `{
"collect":{
	"metric1":{"type": "gauge","labels":["label"]},
	"metric2":{	"type": "counter","labels":["label"]},
	"metric3":{"type": "summary","labels":["label"]},
	"metric4":{"type": "histogram","labels":["label"]}
}
}`
	c := &Config{}
	f := new(bytes.Buffer)
	f.WriteString(cfg)

	err := json.Unmarshal(f.Bytes(), &c)
	if err != nil {
		t.Fatal(err)
	}

	m, err := c.getCollectors()
	assert.NoError(t, err)

	assert.IsType(t, prometheus.NewGaugeVec(prometheus.GaugeOpts{}, []string{}), m["metric1"])
	assert.IsType(t, prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{}), m["metric2"])
	assert.IsType(t, prometheus.NewSummaryVec(prometheus.SummaryOpts{}, []string{}), m["metric3"])
	assert.IsType(t, prometheus.NewHistogramVec(prometheus.HistogramOpts{}, []string{}), m["metric4"])
}
