package metrics

// func Test_Register_RPC_Summary(t *testing.T) {
//	// FOR register method, setup used just to init the rpc
//	client, c := setup(
//		t,
//		`"user_gauge":{
//				"type": "gauge",
//				"labels": ["type", "section"]
//			}`,
//		"6666",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.NoError(t, client.Call("metrics.Declare", &NamedCollector{
//		Name: "custom_summary",
//		Collector: Collector{
//			Namespace: "test_summary",
//			Subsystem: "test_summary",
//			Type:      Summary,
//			Help:      "test_summary",
//			Labels:    nil,
//			Buckets:   nil,
//		},
//	}, &ok))
//	assert.True(t, ok)
//
//	var ok2 bool
//	// Add to custom_summary is not supported
//	assert.Error(t, client.Call("metrics.Add", Metric{
//		Name:   "custom_summary",
//		Value:  100.0,
//		Labels: []string{"type22", "section22"},
//	}, &ok2))
//	// ok should became false
//	assert.False(t, ok2)
//
//	out, _, err := get("http://localhost:6666/metrics")
//	assert.NoError(t, err)
//	assert.Contains(t, out, `test_summary_test_summary_custom_summary_sum 0`)
//	assert.Contains(t, out, `test_summary_test_summary_custom_summary_count 0`)
// }
//
// func Test_Sub_RPC_CollectorError(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_gauge":{
//			    "type": "gauge",
//				"labels": ["type", "section"]
//			}`,
//		"2120",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Sub", Metric{
//		Name:   "user_gauge_2",
//		Value:  100.0,
//		Labels: []string{"missing"},
//	}, &ok))
// }
//
// func Test_Sub_RPC_MetricError(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_gauge":{
//				"type": "gauge",
//				"labels": ["type", "section"]
//			}`,
//		"2121",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Sub", Metric{
//		Name:   "user_gauge",
//		Value:  100.0,
//		Labels: []string{"missing"},
//	}, &ok))
// }
//
// func Test_Sub_RPC_MetricError_2(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_gauge":{
//				"type": "gauge",
//				"labels": ["type", "section"]
//			}`,
//		"2122",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Sub", Metric{
//		Name:  "user_gauge",
//		Value: 100.0,
//	}, &ok))
// }
//
// func Test_Sub_RPC_MetricError_3(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_gauge":{
//				"type": "histogram",
//				"labels": ["type", "section"]
//			}`,
//		"2123",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Sub", Metric{
//		Name:  "user_gauge",
//		Value: 100.0,
//	}, &ok))
// }
//
// // -- observe
//
// func Test_Observe_RPC(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_histogram":{
//				"type": "histogram"
//			}`,
//		"2124",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.NoError(t, client.Call("metrics.Observe", Metric{
//		Name:  "user_histogram",
//		Value: 100.0,
//	}, &ok))
//	assert.True(t, ok)
//
//	out, _, err := get("http://localhost:2124/metrics")
//	assert.NoError(t, err)
//	assert.Contains(t, out, `user_histogram`)
// }
//
// func Test_Observe_RPC_Vector(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_histogram":{
//				"type": "histogram",
//				"labels": ["type", "section"]
//			}`,
//		"2125",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.NoError(t, client.Call("metrics.Observe", Metric{
//		Name:   "user_histogram",
//		Value:  100.0,
//		Labels: []string{"core", "first"},
//	}, &ok))
//	assert.True(t, ok)
//
//	out, _, err := get("http://localhost:2125/metrics")
//	assert.NoError(t, err)
//	assert.Contains(t, out, `user_histogram`)
// }
//
// func Test_Observe_RPC_CollectorError(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_histogram":{
//				"type": "histogram",
//				"labels": ["type", "section"]
//			}`,
//		"2126",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Observe", Metric{
//		Name:   "user_histogram",
//		Value:  100.0,
//		Labels: []string{"missing"},
//	}, &ok))
// }
//
// func Test_Observe_RPC_MetricError(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_histogram":{
//				"type": "histogram",
//				"labels": ["type", "section"]
//			}`,
//		"2127",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Observe", Metric{
//		Name:   "user_histogram",
//		Value:  100.0,
//		Labels: []string{"missing"},
//	}, &ok))
// }
//
// func Test_Observe_RPC_MetricError_2(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_histogram":{
//				"type": "histogram",
//				"labels": ["type", "section"]
//			}`,
//		"2128",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Observe", Metric{
//		Name:  "user_histogram",
//		Value: 100.0,
//	}, &ok))
// }
//
// // -- observe summary
//
// func Test_Observe2_RPC(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_histogram":{
//				"type": "summary"
//			}`,
//		"2129",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.NoError(t, client.Call("metrics.Observe", Metric{
//		Name:  "user_histogram",
//		Value: 100.0,
//	}, &ok))
//	assert.True(t, ok)
//
//	out, _, err := get("http://localhost:2129/metrics")
//	assert.NoError(t, err)
//	assert.Contains(t, out, `user_histogram`)
// }
//
// func Test_Observe2_RPC_Invalid(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_histogram":{
//				"type": "summary"
//			}`,
//		"2130",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Observe", Metric{
//		Name:   "user_histogram_2",
//		Value:  100.0,
//		Labels: []string{"missing"},
//	}, &ok))
// }
//
// func Test_Observe2_RPC_Invalid_2(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_histogram":{
//				"type": "gauge"
//			}`,
//		"2131",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Observe", Metric{
//		Name:  "user_histogram",
//		Value: 100.0,
//	}, &ok))
//}
//
// func Test_Observe2_RPC_Vector(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_histogram":{
//				"type": "summary",
//				"labels": ["type", "section"]
//			}`,
//		"2132",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.NoError(t, client.Call("metrics.Observe", Metric{
//		Name:   "user_histogram",
//		Value:  100.0,
//		Labels: []string{"core", "first"},
//	}, &ok))
//	assert.True(t, ok)
//
//	out, _, err := get("http://localhost:2132/metrics")
//	assert.NoError(t, err)
//	assert.Contains(t, out, `user_histogram`)
// }
//
// func Test_Observe2_RPC_CollectorError(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_histogram":{
//				"type": "summary",
//				"labels": ["type", "section"]
//			}`,
//		"2133",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Observe", Metric{
//		Name:   "user_histogram",
//		Value:  100.0,
//		Labels: []string{"missing"},
//	}, &ok))
// }
//
// func Test_Observe2_RPC_MetricError(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_histogram":{
//				"type": "summary",
//				"labels": ["type", "section"]
//			}`,
//		"2134",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Observe", Metric{
//		Name:   "user_histogram",
//		Value:  100.0,
//		Labels: []string{"missing"},
//	}, &ok))
// }
//
// func Test_Observe2_RPC_MetricError_2(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_histogram":{
//				"type": "summary",
//				"labels": ["type", "section"]
//			}`,
//		"2135",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Observe", Metric{
//		Name:  "user_histogram",
//		Value: 100.0,
//	}, &ok))
// }
//
// // add
// func Test_Add_RPC(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_gauge":{
//				"type": "counter"
//		}`,
//		"2136",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.NoError(t, client.Call("metrics.Add", Metric{
//		Name:  "user_gauge",
//		Value: 100.0,
//	}, &ok))
//	assert.True(t, ok)
//
//	out, _, err := get("http://localhost:2136/metrics")
//	assert.NoError(t, err)
//	assert.Contains(t, out, `user_gauge 100`)
// }
//
// func Test_Add_RPC_Vector(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_gauge":{
//				"type": "counter",
//				"labels": ["type", "section"]
//			}`,
//		"2137",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.NoError(t, client.Call("metrics.Add", Metric{
//		Name:   "user_gauge",
//		Value:  100.0,
//		Labels: []string{"core", "first"},
//	}, &ok))
//	assert.True(t, ok)
//
//	out, _, err := get("http://localhost:2137/metrics")
//	assert.NoError(t, err)
//	assert.Contains(t, out, `user_gauge{section="first",type="core"} 100`)
// }
//
// func Test_Add_RPC_CollectorError(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_gauge":{
//					"type": "counter",
//				"labels": ["type", "section"]
//			}`,
//		"2138",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Add", Metric{
//		Name:   "user_gauge_2",
//		Value:  100.0,
//		Labels: []string{"missing"},
//	}, &ok))
//
//	assert.False(t, ok)
// }
//
// func Test_Add_RPC_MetricError(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_gauge":{
//				"type": "counter",
//				"labels": ["type", "section"]
//			}`,
//		"2139",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Add", Metric{
//		Name:   "user_gauge",
//		Value:  100.0,
//		Labels: []string{"missing"},
//	}, &ok))
//
//	assert.False(t, ok)
// }
//
// func Test_Add_RPC_MetricError_2(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_gauge":{
//				"type": "counter",
//				"labels": ["type", "section"]
//			}`,
//		"2140",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Add", Metric{
//		Name:  "user_gauge",
//		Value: 100.0,
//	}, &ok))
//
//	assert.False(t, ok)
// }
//
// func Test_Add_RPC_MetricError_3(t *testing.T) {
//	client, c := setup(
//		t,
//		`"user_gauge":{
//				"type": "histogram",
//				"labels": ["type", "section"]
//			}`,
//		"2141",
//	)
//	defer c.Stop()
//
//	var ok bool
//	assert.Error(t, client.Call("metrics.Add", Metric{
//		Name:  "user_gauge",
//		Value: 100.0,
//	}, &ok))
// }
