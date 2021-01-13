package metrics

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	json "github.com/json-iterator/go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/rpc"
	"github.com/stretchr/testify/assert"
)

type testCfg struct {
	rpcCfg     string
	metricsCfg string
	target     string
}

func (cfg *testCfg) Get(name string) service.Config {
	if name == ID {
		return &testCfg{target: cfg.metricsCfg}
	}

	if name == rpc.ID {
		return &testCfg{target: cfg.rpcCfg}
	}

	return nil
}

func (cfg *testCfg) Unmarshal(out interface{}) error {
	j := json.ConfigCompatibleWithStandardLibrary
	err := j.Unmarshal([]byte(cfg.target), out)
	return err
}

// get request and return body
func get(url string) (string, *http.Response, error) {
	r, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", nil, err
	}

	err = r.Body.Close()
	if err != nil {
		return "", nil, err
	}
	return string(b), r, err
}

func TestService_Serve(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{metricsCfg: `{
		"address": "localhost:2116"
	}`}))

	s, _ := c.Get(ID)
	assert.NotNil(t, s)

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	out, _, err := get("http://localhost:2116/metrics")
	assert.NoError(t, err)

	assert.Contains(t, out, "go_gc_duration_seconds")
}

func Test_ServiceCustomMetric(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{metricsCfg: `{
		"address": "localhost:2115"
	}`}))

	s, _ := c.Get(ID)
	assert.NotNil(t, s)

	collector := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "my_gauge",
		Help: "My gauge value",
	})

	assert.NoError(t, s.(*Service).Register(collector))

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	collector.Set(100)

	out, _, err := get("http://localhost:2115/metrics")
	assert.NoError(t, err)

	assert.Contains(t, out, "my_gauge 100")
}

func Test_ServiceCustomMetricMust(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{metricsCfg: `{
		"address": "localhost:2114"
	}`}))

	s, _ := c.Get(ID)
	assert.NotNil(t, s)

	collector := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "my_gauge_2",
		Help: "My gauge value",
	})

	s.(*Service).MustRegister(collector)

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	collector.Set(100)

	out, _, err := get("http://localhost:2114/metrics")
	assert.NoError(t, err)

	assert.Contains(t, out, "my_gauge_2 100")
}

func Test_ConfiguredMetric(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{metricsCfg: `{
		"address": "localhost:2113",
		"collect":{
			"user_gauge":{
				"type": "gauge"
			}
		}
	}`}))

	s, _ := c.Get(ID)
	assert.NotNil(t, s)

	assert.True(t, s.(*Service).Enabled())

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	s.(*Service).Collector("user_gauge").(prometheus.Gauge).Set(100)

	assert.Nil(t, s.(*Service).Collector("invalid"))

	out, _, err := get("http://localhost:2113/metrics")
	assert.NoError(t, err)

	assert.Contains(t, out, "user_gauge 100")
}

func Test_ConfiguredDuplicateMetric(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{metricsCfg: `{
		"address": "localhost:2112",
		"collect":{
			"go_gc_duration_seconds":{
				"type": "gauge"
			}
		}
	}`}))

	s, _ := c.Get(ID)
	assert.NotNil(t, s)

	assert.True(t, s.(*Service).Enabled())

	assert.Error(t, c.Serve())
}

func Test_ConfiguredInvalidMetric(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{metricsCfg: `{
		"address": "localhost:2112",
		"collect":{
			"user_gauge":{
				"type": "invalid"
			}
		}

	}`}))

	assert.Error(t, c.Serve())
}
