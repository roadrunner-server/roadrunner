package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

type rpcServer struct{ svc *Service }

type Metric struct {
	// Collector name.
	Name string

	// Collector value.
	Value float64

	// Labels associated with metric. Only for vector metrics.
	Labels []string
}

// Add new metric to the designated collector.
func (rpc *rpcServer) Add(m *Metric, ok *bool) error {
	c := rpc.svc.Collector(m.Name)
	if c == nil {
		return fmt.Errorf("undefined collector `%s`", m.Name)
	}

	switch c.(type) {
	case prometheus.Gauge:
		c.(prometheus.Gauge).Add(m.Value)

	case *prometheus.GaugeVec:
		if len(m.Labels) == 0 {
			return fmt.Errorf("required labels for collector `%s`", m.Name)
		}

		c.(*prometheus.GaugeVec).WithLabelValues(m.Labels...).Add(m.Value)

	case prometheus.Counter:
		c.(prometheus.Counter).Add(m.Value)

	case *prometheus.CounterVec:
		if len(m.Labels) == 0 {
			return fmt.Errorf("required labels for collector `%s`", m.Name)
		}

		c.(*prometheus.CounterVec).WithLabelValues(m.Labels...).Add(m.Value)

	case prometheus.Summary:
		c.(prometheus.Counter).Add(m.Value)

	case *prometheus.SummaryVec:
		if len(m.Labels) == 0 {
			return fmt.Errorf("required labels for collector `%s`", m.Name)
		}

		c.(*prometheus.SummaryVec).WithLabelValues(m.Labels...).Observe(m.Value)

	case prometheus.Histogram:
		c.(prometheus.Histogram).Observe(m.Value)

	case *prometheus.HistogramVec:
		if len(m.Labels) == 0 {
			return fmt.Errorf("required labels for collector `%s`", m.Name)
		}

		c.(*prometheus.HistogramVec).WithLabelValues(m.Labels...).Observe(m.Value)
	}

	*ok = true
	return nil
}

// Set the metric value (only for gaude).
func (rpc *rpcServer) Set(m *Metric, ok *bool) error {
	c := rpc.svc.Collector(m.Name)
	if c == nil {
		return fmt.Errorf("undefined collector `%s`", m.Name)
	}

	switch c.(type) {
	case prometheus.Gauge:
		c.(prometheus.Gauge).Set(m.Value)

	case *prometheus.GaugeVec:
		if len(m.Labels) == 0 {
			return fmt.Errorf("required labels for collector `%s`", m.Name)
		}

		c.(*prometheus.GaugeVec).WithLabelValues(m.Labels...).Set(m.Value)

	default:
		return fmt.Errorf("collector `%s` is not `gauge` type", m.Name)
	}

	*ok = true
	return nil
}
