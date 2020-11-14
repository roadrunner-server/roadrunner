package metrics

// func Test_ServiceCustomMetricMust(t *testing.T) {
//	logger, _ := test.NewNullLogger()
//	logger.SetLevel(logrus.DebugLevel)
//
//	c := service.NewContainer(logger)
//	c.Register(ID, &Plugin{})
//
//	assert.NoError(t, c.Init(&testCfg{metricsCfg: `{
//		"address": "localhost:2114"
//	}`}))
//
//	s, _ := c.Get(ID)
//	assert.NotNil(t, s)
//
//	collector := prometheus.NewGauge(prometheus.GaugeOpts{
//		Name: "my_gauge_2",
//		Help: "My gauge value",
//	})
//
//	s.(*Plugin).MustRegister(collector)
//
//	go func() {
//		err := c.Serve()
//		if err != nil {
//			t.Errorf("error during the Serve: error %v", err)
//		}
//	}()
//	time.Sleep(time.Millisecond * 100)
//	defer c.Stop()
//
//	collector.Set(100)
//
//	out, _, err := get("http://localhost:2114/metrics")
//	assert.NoError(t, err)
//
//	assert.Contains(t, out, "my_gauge_2 100")
// }
//
// func Test_ConfiguredMetric(t *testing.T) {
//	logger, _ := test.NewNullLogger()
//	logger.SetLevel(logrus.DebugLevel)
//
//	c := service.NewContainer(logger)
//	c.Register(ID, &Plugin{})
//
//	assert.NoError(t, c.Init(&testCfg{metricsCfg: `{
//		"address": "localhost:2113",
//		"collect":{
//			"user_gauge":{
//				"type": "gauge"
//			}
//		}
//	}`}))
//
//	s, _ := c.Get(ID)
//	assert.NotNil(t, s)
//
//	assert.True(t, s.(*Plugin).Enabled())
//
//	go func() {
//		err := c.Serve()
//		if err != nil {
//			t.Errorf("error during the Serve: error %v", err)
//		}
//	}()
//	time.Sleep(time.Millisecond * 100)
//	defer c.Stop()
//
//	s.(*Plugin).Collector("user_gauge").(prometheus.Gauge).Set(100)
//
//	assert.Nil(t, s.(*Plugin).Collector("invalid"))
//
//	out, _, err := get("http://localhost:2113/metrics")
//	assert.NoError(t, err)
//
//	assert.Contains(t, out, "user_gauge 100")
// }
//
// func Test_ConfiguredDuplicateMetric(t *testing.T) {
//	logger, _ := test.NewNullLogger()
//	logger.SetLevel(logrus.DebugLevel)
//
//	c := service.NewContainer(logger)
//	c.Register(ID, &Plugin{})
//
//	assert.NoError(t, c.Init(&testCfg{metricsCfg: `{
//		"address": "localhost:2112",
//		"collect":{
//			"go_gc_duration_seconds":{
//				"type": "gauge"
//			}
//		}
//	}`}))
//
//	s, _ := c.Get(ID)
//	assert.NotNil(t, s)
//
//	assert.True(t, s.(*Plugin).Enabled())
//
//	assert.Error(t, c.Serve())
// }
//
// func Test_ConfiguredInvalidMetric(t *testing.T) {
//	logger, _ := test.NewNullLogger()
//	logger.SetLevel(logrus.DebugLevel)
//
//	c := service.NewContainer(logger)
//	c.Register(ID, &Plugin{})
//
//	assert.NoError(t, c.Init(&testCfg{metricsCfg: `{
//		"address": "localhost:2112",
//		"collect":{
//			"user_gauge":{
//				"type": "invalid"
//			}
//		}
//
//	}`}))
//
//	assert.Error(t, c.Serve())
// }
