package metrics

import (
	rpc2 "net/rpc"
	"strconv"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/rpc"
	"github.com/stretchr/testify/assert"
)

var port = 5004

func setup(t *testing.T, metric string, portNum string) (*rpc2.Client, service.Container) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		rpcCfg: `{"enable":true, "listen":"tcp://:` + strconv.Itoa(port) + `"}`,
		metricsCfg: `{
		"address": "localhost:` + portNum + `",
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 200)

	client, err := rs.Client()
	assert.NoError(t, err)
	if err != nil {
		panic(err)
	}

	return client, c
}

func Test_Set_RPC(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "gauge"
		}`,
		"2112",
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
		"2113",
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Set", Metric{
		Name:   "user_gauge",
		Value:  100.0,
		Labels: []string{"core", "first"},
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2113/metrics")
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
		"2114",
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
		"2115",
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
		"2116",
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
		"2117",
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
		"2118",
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

	out, _, err := get("http://localhost:2118/metrics")
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
		"2119",
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

	out, _, err := get("http://localhost:2119/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `user_gauge{section="first",type="core"} 90`)
}

func Test_Register_RPC_Histogram(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "gauge",
				"labels": ["type", "section"]
			}`,
		"2319",
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Declare", &NamedCollector{
		Name: "custom_histogram",
		Collector: Collector{
			Namespace: "test_histogram",
			Subsystem: "test_histogram",
			Type:      Histogram,
			Help:      "test_histogram",
			Labels:    nil,
			Buckets:   []float64{0.1, 0.2, 0.5},
		},
	}, &ok))
	assert.True(t, ok)

	var ok2 bool
	// histogram does not support Add, should be an error
	assert.Error(t, client.Call("metrics.Add", Metric{
		Name: "custom_histogram",
	}, &ok2))
	// ok should became false
	assert.False(t, ok2)

	out, _, err := get("http://localhost:2319/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `TYPE test_histogram_test_histogram_custom_histogram histogram`)

	// check buckets
	assert.Contains(t, out, `test_histogram_test_histogram_custom_histogram_bucket{le="0.1"} 0`)
	assert.Contains(t, out, `test_histogram_test_histogram_custom_histogram_bucket{le="0.2"} 0`)
	assert.Contains(t, out, `test_histogram_test_histogram_custom_histogram_bucket{le="0.5"} 0`)
}

func Test_Register_RPC_Gauge(t *testing.T) {
	// FOR register method, setup used just to init the rpc
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "gauge",
				"labels": ["type", "section"]
			}`,
		"2324",
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Declare", &NamedCollector{
		Name: "custom_gauge",
		Collector: Collector{
			Namespace: "test_gauge",
			Subsystem: "test_gauge",
			Type:      Gauge,
			Help:      "test_gauge",
			Labels:    []string{"type", "section"},
			Buckets:   nil,
		},
	}, &ok))
	assert.True(t, ok)

	var ok2 bool
	// Add to custom_gauge
	assert.NoError(t, client.Call("metrics.Add", Metric{
		Name:   "custom_gauge",
		Value:  100.0,
		Labels: []string{"core", "first"},
	}, &ok2))
	// ok should became true
	assert.True(t, ok2)

	// Subtract from custom runtime metric
	var ok3 bool
	assert.NoError(t, client.Call("metrics.Sub", Metric{
		Name:   "custom_gauge",
		Value:  10.0,
		Labels: []string{"core", "first"},
	}, &ok3))
	assert.True(t, ok3)

	out, _, err := get("http://localhost:2324/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `test_gauge_test_gauge_custom_gauge{section="first",type="core"} 90`)
}

func Test_Register_RPC_Counter(t *testing.T) {
	// FOR register method, setup used just to init the rpc
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "gauge",
				"labels": ["type", "section"]
			}`,
		"2328",
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Declare", &NamedCollector{
		Name: "custom_counter",
		Collector: Collector{
			Namespace: "test_counter",
			Subsystem: "test_counter",
			Type:      Counter,
			Help:      "test_counter",
			Labels:    []string{"type", "section"},
			Buckets:   nil,
		},
	}, &ok))
	assert.True(t, ok)

	var ok2 bool
	// Add to custom_counter
	assert.NoError(t, client.Call("metrics.Add", Metric{
		Name:   "custom_counter",
		Value:  100.0,
		Labels: []string{"type2", "section2"},
	}, &ok2))
	// ok should became true
	assert.True(t, ok2)

	out, _, err := get("http://localhost:2328/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `test_counter_test_counter_custom_counter{section="section2",type="type2"} 100`)
}

func Test_Register_RPC_Summary(t *testing.T) {
	// FOR register method, setup used just to init the rpc
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "gauge",
				"labels": ["type", "section"]
			}`,
		"6666",
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Declare", &NamedCollector{
		Name: "custom_summary",
		Collector: Collector{
			Namespace: "test_summary",
			Subsystem: "test_summary",
			Type:      Summary,
			Help:      "test_summary",
			Labels:    nil,
			Buckets:   nil,
		},
	}, &ok))
	assert.True(t, ok)

	var ok2 bool
	// Add to custom_summary is not supported
	assert.Error(t, client.Call("metrics.Add", Metric{
		Name:   "custom_summary",
		Value:  100.0,
		Labels: []string{"type22", "section22"},
	}, &ok2))
	// ok should became false
	assert.False(t, ok2)

	out, _, err := get("http://localhost:6666/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `test_summary_test_summary_custom_summary_sum 0`)
	assert.Contains(t, out, `test_summary_test_summary_custom_summary_count 0`)
}

func Test_Sub_RPC_CollectorError(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
			    "type": "gauge",
				"labels": ["type", "section"]
			}`,
		"2120",
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
		"2121",
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
		"2122",
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
		"2123",
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
		"2124",
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Observe", Metric{
		Name:  "user_histogram",
		Value: 100.0,
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2124/metrics")
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
		"2125",
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Observe", Metric{
		Name:   "user_histogram",
		Value:  100.0,
		Labels: []string{"core", "first"},
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2125/metrics")
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
		"2126",
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
		"2127",
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
		"2128",
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
		"2129",
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Observe", Metric{
		Name:  "user_histogram",
		Value: 100.0,
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2129/metrics")
	assert.NoError(t, err)
	assert.Contains(t, out, `user_histogram`)
}

func Test_Observe2_RPC_Invalid(t *testing.T) {
	client, c := setup(
		t,
		`"user_histogram":{
				"type": "summary"
			}`,
		"2130",
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
		"2131",
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
		"2132",
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Observe", Metric{
		Name:   "user_histogram",
		Value:  100.0,
		Labels: []string{"core", "first"},
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2132/metrics")
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
		"2133",
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
		"2134",
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
		"2135",
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
		"2136",
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Add", Metric{
		Name:  "user_gauge",
		Value: 100.0,
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2136/metrics")
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
		"2137",
	)
	defer c.Stop()

	var ok bool
	assert.NoError(t, client.Call("metrics.Add", Metric{
		Name:   "user_gauge",
		Value:  100.0,
		Labels: []string{"core", "first"},
	}, &ok))
	assert.True(t, ok)

	out, _, err := get("http://localhost:2137/metrics")
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
		"2138",
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Add", Metric{
		Name:   "user_gauge_2",
		Value:  100.0,
		Labels: []string{"missing"},
	}, &ok))

	assert.False(t, ok)
}

func Test_Add_RPC_MetricError(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "counter",
				"labels": ["type", "section"]
			}`,
		"2139",
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Add", Metric{
		Name:   "user_gauge",
		Value:  100.0,
		Labels: []string{"missing"},
	}, &ok))

	assert.False(t, ok)
}

func Test_Add_RPC_MetricError_2(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "counter",
				"labels": ["type", "section"]
			}`,
		"2140",
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Add", Metric{
		Name:  "user_gauge",
		Value: 100.0,
	}, &ok))

	assert.False(t, ok)
}

func Test_Add_RPC_MetricError_3(t *testing.T) {
	client, c := setup(
		t,
		`"user_gauge":{
				"type": "histogram",
				"labels": ["type", "section"]
			}`,
		"2141",
	)
	defer c.Stop()

	var ok bool
	assert.Error(t, client.Call("metrics.Add", Metric{
		Name:  "user_gauge",
		Value: 100.0,
	}, &ok))
}
