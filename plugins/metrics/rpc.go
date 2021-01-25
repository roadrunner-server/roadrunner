package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type rpcServer struct {
	svc *Plugin
	log logger.Logger
}

// Metric represent single metric produced by the application.
type Metric struct {
	// Collector name.
	Name string

	// Collector value.
	Value float64

	// Labels associated with metric. Only for vector metrics. Must be provided in a form of label values.
	Labels []string
}

// Add new metric to the designated collector.
func (rpc *rpcServer) Add(m *Metric, ok *bool) error {
	const op = errors.Op("metrics_plugin_add")
	rpc.log.Info("adding metric", "name", m.Name, "value", m.Value, "labels", m.Labels)
	c, exist := rpc.svc.collectors.Load(m.Name)
	if !exist {
		rpc.log.Error("undefined collector", "collector", m.Name)
		return errors.E(op, errors.Errorf("undefined collector %s, try first Declare the desired collector", m.Name))
	}

	switch c := c.(type) {
	case prometheus.Gauge:
		c.Add(m.Value)

	case *prometheus.GaugeVec:
		if len(m.Labels) == 0 {
			rpc.log.Error("required labels for collector", "collector", m.Name)
			return errors.E(op, errors.Errorf("required labels for collector %s", m.Name))
		}

		gauge, err := c.GetMetricWithLabelValues(m.Labels...)
		if err != nil {
			rpc.log.Error("failed to get metrics with label values", "collector", m.Name, "labels", m.Labels)
			return errors.E(op, err)
		}
		gauge.Add(m.Value)
	case prometheus.Counter:
		c.Add(m.Value)

	case *prometheus.CounterVec:
		if len(m.Labels) == 0 {
			return errors.E(op, errors.Errorf("required labels for collector `%s`", m.Name))
		}

		gauge, err := c.GetMetricWithLabelValues(m.Labels...)
		if err != nil {
			rpc.log.Error("failed to get metrics with label values", "collector", m.Name, "labels", m.Labels)
			return errors.E(op, err)
		}
		gauge.Add(m.Value)

	default:
		return errors.E(op, errors.Errorf("collector %s does not support method `Add`", m.Name))
	}

	// RPC, set ok to true as return value. Need by rpc.Call reply argument
	*ok = true
	rpc.log.Info("metric successfully added", "name", m.Name, "labels", m.Labels, "value", m.Value)
	return nil
}

// Sub subtract the value from the specific metric (gauge only).
func (rpc *rpcServer) Sub(m *Metric, ok *bool) error {
	const op = errors.Op("metrics_plugin_sub")
	rpc.log.Info("subtracting value from metric", "name", m.Name, "value", m.Value, "labels", m.Labels)
	c, exist := rpc.svc.collectors.Load(m.Name)
	if !exist {
		rpc.log.Error("undefined collector", "name", m.Name, "value", m.Value, "labels", m.Labels)
		return errors.E(op, errors.Errorf("undefined collector %s", m.Name))
	}
	if c == nil {
		// can it be nil ??? I guess can't
		return errors.E(op, errors.Errorf("undefined collector %s", m.Name))
	}

	switch c := c.(type) {
	case prometheus.Gauge:
		c.Sub(m.Value)

	case *prometheus.GaugeVec:
		if len(m.Labels) == 0 {
			rpc.log.Error("required labels for collector, but none was provided", "name", m.Name, "value", m.Value)
			return errors.E(op, errors.Errorf("required labels for collector %s", m.Name))
		}

		gauge, err := c.GetMetricWithLabelValues(m.Labels...)
		if err != nil {
			rpc.log.Error("failed to get metrics with label values", "collector", m.Name, "labels", m.Labels)
			return errors.E(op, err)
		}
		gauge.Sub(m.Value)
	default:
		return errors.E(op, errors.Errorf("collector `%s` does not support method `Sub`", m.Name))
	}
	rpc.log.Info("subtracting operation finished successfully", "name", m.Name, "labels", m.Labels, "value", m.Value)

	*ok = true
	return nil
}

// Observe the value (histogram and summary only).
func (rpc *rpcServer) Observe(m *Metric, ok *bool) error {
	const op = errors.Op("metrics_plugin_observe")
	rpc.log.Info("observing metric", "name", m.Name, "value", m.Value, "labels", m.Labels)

	c, exist := rpc.svc.collectors.Load(m.Name)
	if !exist {
		rpc.log.Error("undefined collector", "name", m.Name, "value", m.Value, "labels", m.Labels)
		return errors.E(op, errors.Errorf("undefined collector %s", m.Name))
	}
	if c == nil {
		return errors.E(op, errors.Errorf("undefined collector %s", m.Name))
	}

	switch c := c.(type) {
	case *prometheus.SummaryVec:
		if len(m.Labels) == 0 {
			return errors.E(op, errors.Errorf("required labels for collector `%s`", m.Name))
		}

		observer, err := c.GetMetricWithLabelValues(m.Labels...)
		if err != nil {
			return errors.E(op, err)
		}
		observer.Observe(m.Value)

	case prometheus.Histogram:
		c.Observe(m.Value)

	case *prometheus.HistogramVec:
		if len(m.Labels) == 0 {
			return errors.E(op, errors.Errorf("required labels for collector `%s`", m.Name))
		}

		observer, err := c.GetMetricWithLabelValues(m.Labels...)
		if err != nil {
			rpc.log.Error("failed to get metrics with label values", "collector", m.Name, "labels", m.Labels)
			return errors.E(op, err)
		}
		observer.Observe(m.Value)
	default:
		return errors.E(op, errors.Errorf("collector `%s` does not support method `Observe`", m.Name))
	}

	rpc.log.Info("observe operation finished successfully", "name", m.Name, "labels", m.Labels, "value", m.Value)

	*ok = true
	return nil
}

// Declare is used to register new collector in prometheus
// THE TYPES ARE:
// 	NamedCollector -> Collector with the name
// 	bool -> RPC reply value
// RETURNS:
// 	error
func (rpc *rpcServer) Declare(nc *NamedCollector, ok *bool) error {
	const op = errors.Op("metrics_plugin_declare")
	rpc.log.Info("declaring new metric", "name", nc.Name, "type", nc.Type, "namespace", nc.Namespace)
	_, exist := rpc.svc.collectors.Load(nc.Name)
	if exist {
		rpc.log.Error("metric with provided name already exist", "name", nc.Name, "type", nc.Type, "namespace", nc.Namespace)
		return errors.E(op, errors.Errorf("tried to register existing collector with the name `%s`", nc.Name))
	}

	var collector prometheus.Collector
	switch nc.Type {
	case Histogram:
		opts := prometheus.HistogramOpts{
			Name:      nc.Name,
			Namespace: nc.Namespace,
			Subsystem: nc.Subsystem,
			Help:      nc.Help,
			Buckets:   nc.Buckets,
		}

		if len(nc.Labels) != 0 {
			collector = prometheus.NewHistogramVec(opts, nc.Labels)
		} else {
			collector = prometheus.NewHistogram(opts)
		}
	case Gauge:
		opts := prometheus.GaugeOpts{
			Name:      nc.Name,
			Namespace: nc.Namespace,
			Subsystem: nc.Subsystem,
			Help:      nc.Help,
		}

		if len(nc.Labels) != 0 {
			collector = prometheus.NewGaugeVec(opts, nc.Labels)
		} else {
			collector = prometheus.NewGauge(opts)
		}
	case Counter:
		opts := prometheus.CounterOpts{
			Name:      nc.Name,
			Namespace: nc.Namespace,
			Subsystem: nc.Subsystem,
			Help:      nc.Help,
		}

		if len(nc.Labels) != 0 {
			collector = prometheus.NewCounterVec(opts, nc.Labels)
		} else {
			collector = prometheus.NewCounter(opts)
		}
	case Summary:
		opts := prometheus.SummaryOpts{
			Name:      nc.Name,
			Namespace: nc.Namespace,
			Subsystem: nc.Subsystem,
			Help:      nc.Help,
		}

		if len(nc.Labels) != 0 {
			collector = prometheus.NewSummaryVec(opts, nc.Labels)
		} else {
			collector = prometheus.NewSummary(opts)
		}

	default:
		return errors.E(op, errors.Errorf("unknown collector type %s", nc.Type))
	}

	// add collector to sync.Map
	rpc.svc.collectors.Store(nc.Name, collector)
	// that method might panic, we handle it by recover
	err := rpc.svc.Register(collector)
	if err != nil {
		*ok = false
		return errors.E(op, err)
	}

	rpc.log.Info("metric successfully added", "name", nc.Name, "type", nc.Type, "namespace", nc.Namespace)

	*ok = true
	return nil
}

// Set the metric value (only for gaude).
func (rpc *rpcServer) Set(m *Metric, ok *bool) (err error) {
	const op = errors.Op("metrics_plugin_set")
	rpc.log.Info("observing metric", "name", m.Name, "value", m.Value, "labels", m.Labels)

	c, exist := rpc.svc.collectors.Load(m.Name)
	if !exist {
		return errors.E(op, errors.Errorf("undefined collector %s", m.Name))
	}
	if c == nil {
		return errors.E(op, errors.Errorf("undefined collector %s", m.Name))
	}

	switch c := c.(type) {
	case prometheus.Gauge:
		c.Set(m.Value)

	case *prometheus.GaugeVec:
		if len(m.Labels) == 0 {
			rpc.log.Error("required labels for collector", "collector", m.Name)
			return errors.E(op, errors.Errorf("required labels for collector %s", m.Name))
		}

		gauge, err := c.GetMetricWithLabelValues(m.Labels...)
		if err != nil {
			rpc.log.Error("failed to get metrics with label values", "collector", m.Name, "labels", m.Labels)
			return errors.E(op, err)
		}
		gauge.Set(m.Value)

	default:
		return errors.E(op, errors.Errorf("collector `%s` does not support method Set", m.Name))
	}

	rpc.log.Info("set operation finished successfully", "name", m.Name, "labels", m.Labels, "value", m.Value)

	*ok = true
	return nil
}
