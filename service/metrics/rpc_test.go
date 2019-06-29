package metrics

import (
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/rpc"
	"github.com/stretchr/testify/assert"
	rpc2 "net/rpc"
	"strconv"
	"testing"
	"time"
)

var port = 5004

func setup(t *testing.T, metric string) (*rpc2.Client, service.Container) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		rpcCfg: `{"enable":true, "listen":"tcp://:` + strconv.Itoa(port) + `"}`,
		metricsCfg: `{
		"address": "localhost:2112",
		"collect":{
			` + metric + `
		}
	}`}))

	// rotate ports for travis
	port++

	s, _ := c.Get(ID)
	assert.NotNil(t, s)

	s2, _ := c.Get(rpc.ID)
	rs := s2.(*rpc.Service)

	assert.True(t, s.(*Service).Enabled())

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)

	client, err := rs.Client()
	assert.NoError(t, err)

	return client, c
}

func Test_Set_RPC(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "gauge"
		}`,
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Set", Metric{
		Name:  "user_gauge",
		Value: 100.0,
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2112/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `user_gauge 100`)
}

func Test_Set_RPC_Vector(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "gauge",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Set", Metric{
		Name:   "user_gauge",
		Value:  100.0,
		Labels: []string{"core", "first"},
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2112/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `user_gauge{section="first",type="core"} 100`)
}

func Test_Set_RPC_CollectorError(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
					"type": "gauge",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Set", Metric{
		Name:   "user_gauge_2",
		Value:  100.0,
		Labels: []string{"missing"},
	}, &ok))
}

func Test_Set_RPC_MetricError(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "gauge",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Set", Metric{
		Name:   "user_gauge",
		Value:  100.0,
		Labels: []string{"missing"},
	}, &ok))
}

func Test_Set_RPC_MetricError_2(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "gauge",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Set", Metric{
		Name:  "user_gauge",
		Value: 100.0,
	}, &ok))
}

func Test_Set_RPC_MetricError_3(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "histogram",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Set", Metric{
		Name:  "user_gauge",
		Value: 100.0,
	}, &ok))
}

// sub

func Test_Sub_RPC(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "gauge"
		}`,
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Add", Metric{
		Name:  "user_gauge",
		Value: 100.0,
	}, &ok))
	assert.True(t, ok)

	assert.NoError(t, client.Call("metrics.Sub", Metric{
		Name:  "user_gauge",
		Value: 10.0,
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2112/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `user_gauge 90`)
}

func Test_Sub_RPC_Vector(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "gauge",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Add", Metric{
		Name:   "user_gauge",
		Value:  100.0,
		Labels: []string{"core", "first"},
	}, &ok))
	assert.True(t, ok)

	assert.NoError(t, client.Call("metrics.Sub", Metric{
		Name:   "user_gauge",
		Value:  10.0,
		Labels: []string{"core", "first"},
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2112/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `user_gauge{section="first",type="core"} 90`)
}

func Test_Sub_RPC_CollectorError(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
			    "type": "gauge",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Sub", Metric{
		Name:   "user_gauge_2",
		Value:  100.0,
		Labels: []string{"missing"},
	}, &ok))
}

func Test_Sub_RPC_MetricError(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "gauge",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Sub", Metric{
		Name:   "user_gauge",
		Value:  100.0,
		Labels: []string{"missing"},
	}, &ok))
}

func Test_Sub_RPC_MetricError_2(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "gauge",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Sub", Metric{
		Name:  "user_gauge",
		Value: 100.0,
	}, &ok))
}

func Test_Sub_RPC_MetricError_3(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "histogram",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Sub", Metric{
		Name:  "user_gauge",
		Value: 100.0,
	}, &ok))
}

// -- observe

func Test_Observe_RPC(t *testing.T) {
	client, c := setup(
		t,
		`"user_histogram":{
				"type": "histogram"
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Observe", Metric{
		Name:  "user_histogram",
		Value: 100.0,
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2112/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `user_histogram`)
}

func Test_Observe_RPC_Vector(t *testing.T) {
	client, c := setup(
		t,
		`"user_histogram":{
				"type": "histogram",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Observe", Metric{
		Name:   "user_histogram",
		Value:  100.0,
		Labels: []string{"core", "first"},
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2112/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `user_histogram`)
}

func Test_Observe_RPC_CollectorError(t *testing.T) {
	client, c := setup(
		t,
		`"user_histogram":{
				"type": "histogram",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Observe", Metric{
		Name:   "user_histogram",
		Value:  100.0,
		Labels: []string{"missing"},
	}, &ok))
}

func Test_Observe_RPC_MetricError(t *testing.T) {
	client, c := setup(
		t,
		`"user_histogram":{
				"type": "histogram",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Observe", Metric{
		Name:   "user_histogram",
		Value:  100.0,
		Labels: []string{"missing"},
	}, &ok))
}

func Test_Observe_RPC_MetricError_2(t *testing.T) {
	client, c := setup(
		t,
		`"user_histogram":{
				"type": "histogram",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Observe", Metric{
		Name:  "user_histogram",
		Value: 100.0,
	}, &ok))
}

// -- observe summary

func Test_Observe2_RPC(t *testing.T) {
	client, c := setup(
		t,
		`"user_histogram":{
				"type": "summary"
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Observe", Metric{
		Name:  "user_histogram",
		Value: 100.0,
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2112/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `user_histogram`)
}

func Test_Observe2_RPC_Invalid(t *testing.T) {
	client, c := setup(
		t,
		`"user_histogram":{
				"type": "summary"
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Observe", Metric{
		Name:   "user_histogram_2",
		Value:  100.0,
		Labels: []string{"missing"},
	}, &ok))
}

func Test_Observe2_RPC_Invalid_2(t *testing.T) {
	client, c := setup(
		t,
		`"user_histogram":{
				"type": "gauge"
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Observe", Metric{
		Name:  "user_histogram",
		Value: 100.0,
	}, &ok))
}

func Test_Observe2_RPC_Vector(t *testing.T) {
	client, c := setup(
		t,
		`"user_histogram":{
				"type": "summary",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Observe", Metric{
		Name:   "user_histogram",
		Value:  100.0,
		Labels: []string{"core", "first"},
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2112/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `user_histogram`)
}

func Test_Observe2_RPC_CollectorError(t *testing.T) {
	client, c := setup(
		t,
		`"user_histogram":{
				"type": "summary",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Observe", Metric{
		Name:   "user_histogram",
		Value:  100.0,
		Labels: []string{"missing"},
	}, &ok))
}

func Test_Observe2_RPC_MetricError(t *testing.T) {
	client, c := setup(
		t,
		`"user_histogram":{
				"type": "summary",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Observe", Metric{
		Name:   "user_histogram",
		Value:  100.0,
		Labels: []string{"missing"},
	}, &ok))
}

func Test_Observe2_RPC_MetricError_2(t *testing.T) {
	client, c := setup(
		t,
		`"user_histogram":{
				"type": "summary",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Observe", Metric{
		Name:  "user_histogram",
		Value: 100.0,
	}, &ok))
}

// add
func Test_Add_RPC(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "counter"
		}`,
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Add", Metric{
		Name:  "user_gauge",
		Value: 100.0,
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2112/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `user_gauge 100`)
}

func Test_Add_RPC_Vector(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "counter",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Add", Metric{
		Name:   "user_gauge",
		Value:  100.0,
		Labels: []string{"core", "first"},
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2112/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `user_gauge{section="first",type="core"} 100`)
}

func Test_Add_RPC_CollectorError(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
					"type": "counter",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Add", Metric{
		Name:   "user_gauge_2",
		Value:  100.0,
		Labels: []string{"missing"},
	}, &ok))
}

func Test_Add_RPC_MetricError(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "counter",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Add", Metric{
		Name:   "user_gauge",
		Value:  100.0,
		Labels: []string{"missing"},
	}, &ok))
}

func Test_Add_RPC_MetricError_2(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "counter",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Add", Metric{
		Name:  "user_gauge",
		Value: 100.0,
	}, &ok))
}

func Test_Add_RPC_MetricError_3(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "histogram",
				"labels": ["type", "section"]
			}`,
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Add", Metric{
		Name:  "user_gauge",
		Value: 100.0,
	}, &ok))
}
