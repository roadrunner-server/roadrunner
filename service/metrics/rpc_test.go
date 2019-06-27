package metrics

import (
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/rpc"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_Set_RPC(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		rpcCfg: `{"enable":true, "listen":"tcp://:5004"}`,
		metricsCfg: `{
		"address": "localhost:2112",
		"collect":{
			"user_gauge":{
				"type": "gauge"
			}
		}

	}`}))

	s, _ := c.Get(ID)
	assert.NotNil(t, s)

	s2, _ := c.Get(rpc.ID)
	rs := s2.(*rpc.Service)

	assert.True(t, s.(*Service).Enabled())

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	client, err := rs.Client()
	assert.NoError(t, err)

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
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		rpcCfg: `{"enable":true, "listen":"tcp://:5004"}`,
		metricsCfg: `{
		"address": "localhost:2112",
		"collect":{
			"user_gauge":{
				"type": "gauge",
				"labels": ["type", "section"]
			}
		}

	}`}))

	s, _ := c.Get(ID)
	assert.NotNil(t, s)

	s2, _ := c.Get(rpc.ID)
	rs := s2.(*rpc.Service)

	assert.True(t, s.(*Service).Enabled())

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	client, err := rs.Client()
	assert.NoError(t, err)

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
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		rpcCfg: `{"enable":true, "listen":"tcp://:5004"}`,
		metricsCfg: `{
		"address": "localhost:2112",
		"collect":{
			"user_gauge":{
				"type": "gauge",
				"labels": ["type", "section"]
			}
		}
	}`}))

	s, _ := c.Get(ID)
	assert.NotNil(t, s)

	s2, _ := c.Get(rpc.ID)
	rs := s2.(*rpc.Service)

	assert.True(t, s.(*Service).Enabled())

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	client, err := rs.Client()
	assert.NoError(t, err)

	var ok bool
	assert.Error(t, client.Call("metrics.Set", Metric{
		Name:   "user_gauge_2",
		Value:  100.0,
		Labels: []string{"missing"},
	}, &ok))
}

func Test_Set_RPC_MetricError(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		rpcCfg: `{"enable":true, "listen":"tcp://:5004"}`,
		metricsCfg: `{
		"address": "localhost:2112",
		"collect":{
			"user_gauge":{
				"type": "gauge",
				"labels": ["type", "section"]
			}
		}

	}`}))

	s, _ := c.Get(ID)
	assert.NotNil(t, s)

	s2, _ := c.Get(rpc.ID)
	rs := s2.(*rpc.Service)

	assert.True(t, s.(*Service).Enabled())

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	client, err := rs.Client()
	assert.NoError(t, err)

	var ok bool
	assert.Error(t, client.Call("metrics.Set", Metric{
		Name:   "user_gauge",
		Value:  100.0,
		Labels: []string{"missing"},
	}, &ok))
}
