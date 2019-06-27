package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

type rpcServer struct{ svc *Service }

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
func (rpc *rpcServer) Add(m *Metric, ok *bool) (err error) {
	defer func() {
		if r, fail := recover().(error); fail {
			err = r
		}
	}()

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

	default:
		return fmt.Errorf("collector `%s` does not support method `Add`", m.Name)
	}

	*ok = true
	return nil
}

// Sub subtract the value from the specific metric (gauge only).
func (rpc *rpcServer) Sub(m *Metric, ok *bool) (err error) {
	defer func() {
		if r, fail := recover().(error); fail {
			err = r
		}
	}()

	c := rpc.svc.Collector(m.Name)
	if c == nil {
		return fmt.Errorf("undefined collector `%s`", m.Name)
	}

	switch c.(type) {
	case prometheus.Gauge:
		c.(prometheus.Gauge).Sub(m.Value)

	case *prometheus.GaugeVec:
		if len(m.Labels) == 0 {
			return fmt.Errorf("required labels for collector `%s`", m.Name)
		}

		c.(*prometheus.GaugeVec).WithLabelValues(m.Labels...).Sub(m.Value)
	default:
		return fmt.Errorf("collector `%s` does not support method `Sub`", m.Name)
	}

	*ok = true
	return nil
}

// Observe the value (histogram and summary only).
func (rpc *rpcServer) Observe(m *Metric, ok *bool) (err error) {
	defer func() {
		if r, fail := recover().(error); fail {
			err = r
		}
	}()

	c := rpc.svc.Collector(m.Name)
	if c == nil {
		return fmt.Errorf("undefined collector `%s`", m.Name)
	}

	switch c.(type) {
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
	default:
		return fmt.Errorf("collector `%s` does not support method `Observe`", m.Name)
	}

	*ok = true
	return nil
}

// Set the metric value (only for gaude).
func (rpc *rpcServer) Set(m *Metric, ok *bool) (err error) {
	defer func() {
		if r, fail := recover().(error); fail {
			err = r
		}
	}()

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
		return fmt.Errorf("collector `%s` does not support method `Set`", m.Name)
	}

	*ok = true
	return nil
}
