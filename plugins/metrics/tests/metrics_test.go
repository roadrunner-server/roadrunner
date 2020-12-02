package tests

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/spiral/endure"
	"github.com/spiral/goridge/v2"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/metrics"
	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/stretchr/testify/assert"
)

const dialAddr = "127.0.0.1:6001"
const dialNetwork = "tcp"
const getAddr = "http://localhost:2112/metrics"

// get request and return body
func get() (string, error) {
	r, err := http.Get(getAddr)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	err = r.Body.Close()
	if err != nil {
		return "", err
	}
	// unsafe
	return string(b), err
}

func TestMetricsInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Viper{}
	cfg.Prefix = "rr"
	cfg.Path = ".rr-test.yaml"

	err = cont.RegisterAll(
		cfg,
		&metrics.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&Plugin1{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	tt := time.NewTimer(time.Second * 5)

	out, err := get()
	assert.NoError(t, err)

	assert.Contains(t, out, "go_gc_duration_seconds")

	for {
		select {
		case e := <-ch:
			assert.Fail(t, "error", e.Error.Error())
			err = cont.Stop()
			if err != nil {
				assert.FailNow(t, "error", err.Error())
			}
		case <-sig:
			err = cont.Stop()
			if err != nil {
				assert.FailNow(t, "error", err.Error())
			}
			return
		case <-tt.C:
			// timeout
			err = cont.Stop()
			if err != nil {
				assert.FailNow(t, "error", err.Error())
			}
			return
		}
	}
}

func TestMetricsGaugeCollector(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Viper{}
	cfg.Prefix = "rr"
	cfg.Path = ".rr-test.yaml"

	err = cont.RegisterAll(
		cfg,
		&metrics.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&Plugin1{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	time.Sleep(time.Second)
	tt := time.NewTimer(time.Second * 5)

	out, err := get()
	assert.NoError(t, err)
	assert.Contains(t, out, "my_gauge 100")
	assert.Contains(t, out, "my_gauge2 100")

	out, err = get()
	assert.NoError(t, err)
	assert.Contains(t, out, "go_gc_duration_seconds")

	for {
		select {
		case e := <-ch:
			assert.Fail(t, "error", e.Error.Error())
			err = cont.Stop()
			if err != nil {
				assert.FailNow(t, "error", err.Error())
			}
		case <-sig:
			err = cont.Stop()
			if err != nil {
				assert.FailNow(t, "error", err.Error())
			}
			return
		case <-tt.C:
			// timeout
			err = cont.Stop()
			if err != nil {
				assert.FailNow(t, "error", err.Error())
			}
			return
		}
	}
}

func TestMetricsDifferentRPCCalls(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Viper{}
	cfg.Prefix = "rr"
	cfg.Path = ".rr-test.yaml"

	err = cont.RegisterAll(
		cfg,
		&metrics.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		tt := time.NewTimer(time.Minute * 3)
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-tt.C:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	t.Run("DeclareMetric", declareMetricsTest)
	genericOut, err := get()
	assert.NoError(t, err)
	assert.Contains(t, genericOut, "test_metrics_named_collector")

	t.Run("AddMetric", addMetricsTest)
	genericOut, err = get()
	assert.NoError(t, err)
	assert.Contains(t, genericOut, "test_metrics_named_collector 10000")

	t.Run("SetMetric", setMetric)
	genericOut, err = get()
	assert.NoError(t, err)
	assert.Contains(t, genericOut, "user_gauge_collector 100")

	t.Run("VectorMetric", vectorMetric)
	genericOut, err = get()
	assert.NoError(t, err)
	assert.Contains(t, genericOut, "gauge_2_collector{section=\"first\",type=\"core\"} 100")

	t.Run("MissingSection", missingSection)
	t.Run("SetWithoutLabels", setWithoutLabels)
	t.Run("SetOnHistogram", setOnHistogram)
	t.Run("MetricSub", subMetric)
	genericOut, err = get()
	assert.NoError(t, err)
	assert.Contains(t, genericOut, "sub_gauge_subMetric 1")

	t.Run("SubVector", subVector)
	genericOut, err = get()
	assert.NoError(t, err)
	assert.Contains(t, genericOut, "sub_gauge_subVector{section=\"first\",type=\"core\"} 1")

	t.Run("RegisterHistogram", registerHistogram)

	genericOut, err = get()
	assert.NoError(t, err)
	assert.Contains(t, genericOut, `TYPE histogram_registerHistogram`)

	// check buckets
	assert.Contains(t, genericOut, `histogram_registerHistogram_bucket{le="0.1"} 0`)
	assert.Contains(t, genericOut, `histogram_registerHistogram_bucket{le="0.2"} 0`)
	assert.Contains(t, genericOut, `histogram_registerHistogram_bucket{le="0.5"} 0`)
	assert.Contains(t, genericOut, `histogram_registerHistogram_bucket{le="+Inf"} 0`)
	assert.Contains(t, genericOut, `histogram_registerHistogram_sum 0`)
	assert.Contains(t, genericOut, `histogram_registerHistogram_count 0`)

	t.Run("CounterMetric", counterMetric)
	genericOut, err = get()
	assert.NoError(t, err)
	assert.Contains(t, genericOut, "HELP default_default_counter_CounterMetric test_counter")
	assert.Contains(t, genericOut, `default_default_counter_CounterMetric{section="section2",type="type2"}`)

	t.Run("ObserveMetric", observeMetric)
	genericOut, err = get()
	assert.NoError(t, err)
	assert.Contains(t, genericOut, "observe_observeMetric")

	t.Run("ObserveMetricNotEnoughLabels", observeMetricNotEnoughLabels)

	close(sig)
}

func observeMetricNotEnoughLabels(t *testing.T) {
	conn, err := net.Dial(dialNetwork, dialAddr)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()
	client := rpc.NewClientWithCodec(goridge.NewClientCodec(conn))
	var ret bool

	nc := metrics.NamedCollector{
		Name: "observe_observeMetricNotEnoughLabels",
		Collector: metrics.Collector{
			Namespace: "default",
			Subsystem: "default",
			Help:      "test_observe",
			Type:      metrics.Histogram,
			Labels:    []string{"type", "section"},
		},
	}

	err = client.Call("metrics.Declare", nc, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)
	ret = false

	assert.Error(t, client.Call("metrics.Observe", metrics.Metric{
		Name:   "observe_observeMetric",
		Value:  100.0,
		Labels: []string{"test"},
	}, &ret))
	assert.False(t, ret)
}

func observeMetric(t *testing.T) {
	conn, err := net.Dial(dialNetwork, dialAddr)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()
	client := rpc.NewClientWithCodec(goridge.NewClientCodec(conn))
	var ret bool

	nc := metrics.NamedCollector{
		Name: "observe_observeMetric",
		Collector: metrics.Collector{
			Namespace: "default",
			Subsystem: "default",
			Help:      "test_observe",
			Type:      metrics.Histogram,
			Labels:    []string{"type", "section"},
		},
	}

	err = client.Call("metrics.Declare", nc, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)
	ret = false

	assert.NoError(t, client.Call("metrics.Observe", metrics.Metric{
		Name:   "observe_observeMetric",
		Value:  100.0,
		Labels: []string{"test", "test2"},
	}, &ret))
	assert.True(t, ret)
}

func counterMetric(t *testing.T) {
	conn, err := net.Dial(dialNetwork, dialAddr)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()
	client := rpc.NewClientWithCodec(goridge.NewClientCodec(conn))
	var ret bool

	nc := metrics.NamedCollector{
		Name: "counter_CounterMetric",
		Collector: metrics.Collector{
			Namespace: "default",
			Subsystem: "default",
			Help:      "test_counter",
			Type:      metrics.Counter,
			Labels:    []string{"type", "section"},
		},
	}

	err = client.Call("metrics.Declare", nc, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)

	ret = false

	assert.NoError(t, client.Call("metrics.Add", metrics.Metric{
		Name:   "counter_CounterMetric",
		Value:  100.0,
		Labels: []string{"type2", "section2"},
	}, &ret))
	assert.True(t, ret)
}

func registerHistogram(t *testing.T) {
	conn, err := net.Dial(dialNetwork, dialAddr)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()
	client := rpc.NewClientWithCodec(goridge.NewClientCodec(conn))
	var ret bool

	nc := metrics.NamedCollector{
		Name: "histogram_registerHistogram",
		Collector: metrics.Collector{
			Help:    "test_histogram",
			Type:    metrics.Histogram,
			Buckets: []float64{0.1, 0.2, 0.5},
		},
	}

	err = client.Call("metrics.Declare", nc, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)

	ret = false

	m := metrics.Metric{
		Name:   "histogram_registerHistogram",
		Value:  10000,
		Labels: nil,
	}

	err = client.Call("metrics.Add", m, &ret)
	assert.Error(t, err)
	assert.False(t, ret)
}

func subVector(t *testing.T) {
	conn, err := net.Dial(dialNetwork, dialAddr)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()

	client := rpc.NewClientWithCodec(goridge.NewClientCodec(conn))
	var ret bool

	nc := metrics.NamedCollector{
		Name: "sub_gauge_subVector",
		Collector: metrics.Collector{
			Namespace: "default",
			Subsystem: "default",
			Type:      metrics.Gauge,
			Labels:    []string{"type", "section"},
		},
	}

	err = client.Call("metrics.Declare", nc, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)
	ret = false

	m := metrics.Metric{
		Name:   "sub_gauge_subVector",
		Value:  100000,
		Labels: []string{"core", "first"},
	}

	err = client.Call("metrics.Add", m, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)
	ret = false

	m = metrics.Metric{
		Name:   "sub_gauge_subVector",
		Value:  99999,
		Labels: []string{"core", "first"},
	}

	err = client.Call("metrics.Sub", m, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)
}

func subMetric(t *testing.T) {
	conn, err := net.Dial(dialNetwork, dialAddr)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()
	client := rpc.NewClientWithCodec(goridge.NewClientCodec(conn))
	var ret bool

	nc := metrics.NamedCollector{
		Name: "sub_gauge_subMetric",
		Collector: metrics.Collector{
			Namespace: "default",
			Subsystem: "default",
			Type:      metrics.Gauge,
		},
	}

	err = client.Call("metrics.Declare", nc, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)
	ret = false

	m := metrics.Metric{
		Name:  "sub_gauge_subMetric",
		Value: 100000,
	}

	err = client.Call("metrics.Add", m, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)
	ret = false

	m = metrics.Metric{
		Name:  "sub_gauge_subMetric",
		Value: 99999,
	}

	err = client.Call("metrics.Sub", m, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)
}

func setOnHistogram(t *testing.T) {
	conn, err := net.Dial(dialNetwork, dialAddr)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()
	client := rpc.NewClientWithCodec(goridge.NewClientCodec(conn))
	var ret bool

	nc := metrics.NamedCollector{
		Name: "histogram_setOnHistogram",
		Collector: metrics.Collector{
			Namespace: "default",
			Subsystem: "default",
			Type:      metrics.Histogram,
			Labels:    []string{"type", "section"},
		},
	}

	err = client.Call("metrics.Declare", nc, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)

	ret = false

	m := metrics.Metric{
		Name:  "gauge_setOnHistogram",
		Value: 100.0,
	}

	err = client.Call("metrics.Set", m, &ret) // expected 2 label values but got 1 in []string{"missing"}
	assert.Error(t, err)
	assert.False(t, ret)
}

func setWithoutLabels(t *testing.T) {
	conn, err := net.Dial(dialNetwork, dialAddr)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()
	client := rpc.NewClientWithCodec(goridge.NewClientCodec(conn))
	var ret bool

	nc := metrics.NamedCollector{
		Name: "gauge_setWithoutLabels",
		Collector: metrics.Collector{
			Namespace: "default",
			Subsystem: "default",
			Type:      metrics.Gauge,
			Labels:    []string{"type", "section"},
		},
	}

	err = client.Call("metrics.Declare", nc, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)

	ret = false

	m := metrics.Metric{
		Name:  "gauge_setWithoutLabels",
		Value: 100.0,
	}

	err = client.Call("metrics.Set", m, &ret) // expected 2 label values but got 1 in []string{"missing"}
	assert.Error(t, err)
	assert.False(t, ret)
}

func missingSection(t *testing.T) {
	conn, err := net.Dial(dialNetwork, dialAddr)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()
	client := rpc.NewClientWithCodec(goridge.NewClientCodec(conn))
	var ret bool

	nc := metrics.NamedCollector{
		Name: "gauge_missing_section_collector",
		Collector: metrics.Collector{
			Namespace: "default",
			Subsystem: "default",
			Type:      metrics.Gauge,
			Labels:    []string{"type", "section"},
		},
	}

	err = client.Call("metrics.Declare", nc, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)

	ret = false

	m := metrics.Metric{
		Name:   "gauge_missing_section_collector",
		Value:  100.0,
		Labels: []string{"missing"},
	}

	err = client.Call("metrics.Set", m, &ret) // expected 2 label values but got 1 in []string{"missing"}
	assert.Error(t, err)
	assert.False(t, ret)
}

func vectorMetric(t *testing.T) {
	conn, err := net.Dial(dialNetwork, dialAddr)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()
	client := rpc.NewClientWithCodec(goridge.NewClientCodec(conn))
	var ret bool

	nc := metrics.NamedCollector{
		Name: "gauge_2_collector",
		Collector: metrics.Collector{
			Namespace: "default",
			Subsystem: "default",
			Type:      metrics.Gauge,
			Labels:    []string{"type", "section"},
		},
	}

	err = client.Call("metrics.Declare", nc, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)

	ret = false

	m := metrics.Metric{
		Name:   "gauge_2_collector",
		Value:  100.0,
		Labels: []string{"core", "first"},
	}

	err = client.Call("metrics.Set", m, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)
}

func setMetric(t *testing.T) {
	conn, err := net.Dial(dialNetwork, dialAddr)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()
	client := rpc.NewClientWithCodec(goridge.NewClientCodec(conn))
	var ret bool

	nc := metrics.NamedCollector{
		Name: "user_gauge_collector",
		Collector: metrics.Collector{
			Namespace: "default",
			Subsystem: "default",
			Type:      metrics.Gauge,
		},
	}

	err = client.Call("metrics.Declare", nc, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)
	ret = false

	m := metrics.Metric{
		Name:  "user_gauge_collector",
		Value: 100.0,
	}

	err = client.Call("metrics.Set", m, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)
}

func addMetricsTest(t *testing.T) {
	conn, err := net.Dial(dialNetwork, dialAddr)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()
	client := rpc.NewClientWithCodec(goridge.NewClientCodec(conn))
	var ret bool

	m := metrics.Metric{
		Name:   "test_metrics_named_collector",
		Value:  10000,
		Labels: nil,
	}

	err = client.Call("metrics.Add", m, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)
}

func declareMetricsTest(t *testing.T) {
	conn, err := net.Dial(dialNetwork, dialAddr)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()
	client := rpc.NewClientWithCodec(goridge.NewClientCodec(conn))
	var ret bool

	nc := metrics.NamedCollector{
		Name: "test_metrics_named_collector",
		Collector: metrics.Collector{
			Namespace: "default",
			Subsystem: "default",
			Type:      metrics.Counter,
			Help:      "NO HELP!",
			Labels:    nil,
			Buckets:   nil,
		},
	}

	err = client.Call("metrics.Declare", nc, &ret)
	assert.NoError(t, err)
	assert.True(t, ret)
}
