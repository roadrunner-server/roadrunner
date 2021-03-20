package metrics

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

	"github.com/golang/mock/gomock"
	endure "github.com/spiral/endure/pkg/container"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/metrics"
	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/spiral/roadrunner/v2/tests/mocks"
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
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
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
	defer tt.Stop()

	out, err := get()
	assert.NoError(t, err)

	assert.Contains(t, out, "go_gc_duration_seconds")
	assert.Contains(t, out, "app_metric_counter")

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
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
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
	defer tt.Stop()

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
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Viper{}
	cfg.Prefix = "rr"
	cfg.Path = ".rr-test.yaml"

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	mockLogger.EXPECT().Debug("worker destructed", "pid", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("worker constructed", "pid", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("Started RPC service", "address", "tcp://127.0.0.1:6001", "services", []string{"metrics"}).MinTimes(1)

	mockLogger.EXPECT().Info("adding metric", "name", "counter_CounterMetric", "value", gomock.Any(), "labels", []string{"type2", "section2"}).MinTimes(1)
	mockLogger.EXPECT().Info("adding metric", "name", "histogram_registerHistogram", "value", gomock.Any(), "labels", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("adding metric", "name", "sub_gauge_subVector", "value", gomock.Any(), "labels", []string{"core", "first"}).MinTimes(1)
	mockLogger.EXPECT().Info("adding metric", "name", "sub_gauge_subMetric", "value", gomock.Any(), "labels", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("adding metric", "name", "test_metrics_named_collector", "value", gomock.Any(), "labels", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("adding metric", "name", "app_metric_counter", "value", gomock.Any(), "labels", gomock.Any()).MinTimes(1)

	mockLogger.EXPECT().Info("metric successfully added", "name", "observe_observeMetricNotEnoughLabels", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "observe_observeMetric", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "counter_CounterMetric", "labels", []string{"type2", "section2"}, "value", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "counter_CounterMetric", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "histogram_registerHistogram", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "sub_gauge_subVector", "labels", []string{"core", "first"}, "value", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "sub_gauge_subVector", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "sub_gauge_subMetric", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "sub_gauge_subMetric", "labels", gomock.Any(), "value", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "histogram_setOnHistogram", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "gauge_setWithoutLabels", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "gauge_missing_section_collector", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "gauge_2_collector", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "test_metrics_named_collector", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "test_metrics_named_collector", "labels", gomock.Any(), "value", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "user_gauge_collector", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("metric successfully added", "name", "app_metric_counter", "labels", gomock.Any(), "value", gomock.Any()).MinTimes(1)

	mockLogger.EXPECT().Info("declaring new metric", "name", "observe_observeMetricNotEnoughLabels", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("declaring new metric", "name", "observe_observeMetric", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("declaring new metric", "name", "counter_CounterMetric", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("declaring new metric", "name", "histogram_registerHistogram", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("declaring new metric", "name", "sub_gauge_subVector", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("declaring new metric", "name", "sub_gauge_subMetric", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("declaring new metric", "name", "histogram_setOnHistogram", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("declaring new metric", "name", "gauge_setWithoutLabels", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("declaring new metric", "name", "gauge_missing_section_collector", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("declaring new metric", "name", "test_metrics_named_collector", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("declaring new metric", "name", "gauge_2_collector", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("declaring new metric", "name", "user_gauge_collector", "type", gomock.Any(), "namespace", gomock.Any()).MinTimes(1)

	mockLogger.EXPECT().Info("observing metric", "name", "observe_observeMetric", "value", gomock.Any(), "labels", []string{"test"}).MinTimes(1)
	mockLogger.EXPECT().Info("observing metric", "name", "observe_observeMetric", "value", gomock.Any(), "labels", []string{"test", "test2"}).MinTimes(1)
	mockLogger.EXPECT().Info("observing metric", "name", "gauge_setOnHistogram", "value", gomock.Any(), "labels", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("observing metric", "name", "gauge_setWithoutLabels", "value", gomock.Any(), "labels", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("observing metric", "name", "gauge_missing_section_collector", "value", gomock.Any(), "labels", []string{"missing"}).MinTimes(1)
	mockLogger.EXPECT().Info("observing metric", "name", "user_gauge_collector", "value", gomock.Any(), "labels", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("observing metric", "name", "gauge_2_collector", "value", gomock.Any(), "labels", []string{"core", "first"}).MinTimes(1)

	mockLogger.EXPECT().Info("observe operation finished successfully", "name", "observe_observeMetric", "labels", []string{"test", "test2"}, "value", gomock.Any()).MinTimes(1)

	mockLogger.EXPECT().Info("set operation finished successfully", "name", "gauge_2_collector", "labels", []string{"core", "first"}, "value", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("set operation finished successfully", "name", "user_gauge_collector", "labels", gomock.Any(), "value", gomock.Any()).MinTimes(1)

	mockLogger.EXPECT().Info("subtracting value from metric", "name", "sub_gauge_subVector", "value", gomock.Any(), "labels", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("subtracting value from metric", "name", "sub_gauge_subMetric", "value", gomock.Any(), "labels", gomock.Any()).MinTimes(1)

	mockLogger.EXPECT().Info("subtracting operation finished successfully", "name", "sub_gauge_subVector", "labels", gomock.Any(), "value", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("subtracting operation finished successfully", "name", "sub_gauge_subMetric", "labels", gomock.Any(), "value", gomock.Any()).MinTimes(1)

	mockLogger.EXPECT().Error("failed to get metrics with label values", "collector", "gauge_missing_section_collector", "labels", []string{"missing"}).MinTimes(1)
	mockLogger.EXPECT().Error("required labels for collector", "collector", "gauge_setWithoutLabels").MinTimes(1)
	mockLogger.EXPECT().Error("failed to get metrics with label values", "collector", "observe_observeMetric", "labels", []string{"test"}).MinTimes(1)

	err = cont.RegisterAll(
		cfg,
		&metrics.Plugin{},
		&rpcPlugin.Plugin{},
		mockLogger,
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

	tt := time.NewTimer(time.Minute * 3)
	defer tt.Stop()

	go func() {
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

	t.Run("ConfiguredCounterMetric", configuredCounterMetric)
	genericOut, err = get()
	assert.NoError(t, err)
	assert.Contains(t, genericOut, "HELP app_metric_counter Custom application counter.")
	assert.Contains(t, genericOut, `app_metric_counter 100`)

	close(sig)
}

func configuredCounterMetric(t *testing.T) {
	conn, err := net.Dial(dialNetwork, dialAddr)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	var ret bool

	assert.NoError(t, client.Call("metrics.Add", metrics.Metric{
		Name:  "app_metric_counter",
		Value: 100.0,
	}, &ret))
	assert.True(t, ret)
}

func observeMetricNotEnoughLabels(t *testing.T) {
	conn, err := net.Dial(dialNetwork, dialAddr)
	assert.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
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
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
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
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
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
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
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

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
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
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
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
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
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
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
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
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
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
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
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
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
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
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
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
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
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
