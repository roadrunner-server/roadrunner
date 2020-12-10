package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type rpcServer struct {
	svc *Service
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

	switch c := c.(type) {
	case prometheus.Gauge:
		c.Add(m.Value)

	case *prometheus.GaugeVec:
		if len(m.Labels) == 0 {
			return fmt.Errorf("required labels for collector `%s`", m.Name)
		}

		c.WithLabelValues(m.Labels...).Add(m.Value)

	case prometheus.Counter:
		c.Add(m.Value)

	case *prometheus.CounterVec:
		if len(m.Labels) == 0 {
			return fmt.Errorf("required labels for collector `%s`", m.Name)
		}

		c.WithLabelValues(m.Labels...).Add(m.Value)

	default:
		return fmt.Errorf("collector `%s` does not support method `Add`", m.Name)
	}

	// RPC, set ok to true as return value. Need by rpc.Call reply argument
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

	switch c := c.(type) {
	case prometheus.Gauge:
		c.Sub(m.Value)

	case *prometheus.GaugeVec:
		if len(m.Labels) == 0 {
			return fmt.Errorf("required labels for collector `%s`", m.Name)
		}

		c.WithLabelValues(m.Labels...).Sub(m.Value)
	default:
		return fmt.Errorf("collector `%s` does not support method `Sub`", m.Name)
	}

	// RPC, set ok to true as return value. Need by rpc.Call reply argument
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

	switch c := c.(type) {
	case *prometheus.SummaryVec:
		if len(m.Labels) == 0 {
			return fmt.Errorf("required labels for collector `%s`", m.Name)
		}

		c.WithLabelValues(m.Labels...).Observe(m.Value)

	case prometheus.Histogram:
		c.Observe(m.Value)

	case *prometheus.HistogramVec:
		if len(m.Labels) == 0 {
			return fmt.Errorf("required labels for collector `%s`", m.Name)
		}

		c.WithLabelValues(m.Labels...).Observe(m.Value)
	default:
		return fmt.Errorf("collector `%s` does not support method `Observe`", m.Name)
	}

	// RPC, set ok to true as return value. Need by rpc.Call reply argument
	*ok = true
	return nil
}

// Declare is used to register new collector in prometheus
// THE TYPES ARE:
// 	NamedCollector -> Collector with the name
// 	bool -> RPC reply value
// RETURNS:
// 	error
func (rpc *rpcServer) Declare(c *NamedCollector, ok *bool) (err error) {
	// MustRegister could panic, so, to return error and not shutdown whole app
	// we recover and return error
	defer func() {
		if r, fail := recover().(error); fail {
			err = r
		}
	}()

	if rpc.svc.Collector(c.Name) != nil {
		*ok = false
		// alternative is to return error
		// fmt.Errorf("tried to register existing collector with the name `%s`", c.Name)
		return nil
	}

	var collector prometheus.Collector
	switch c.Type {
	case Histogram:
		opts := prometheus.HistogramOpts{
			Name:      c.Name,
			Namespace: c.Namespace,
			Subsystem: c.Subsystem,
			Help:      c.Help,
			Buckets:   c.Buckets,
		}

		if len(c.Labels) != 0 {
			collector = prometheus.NewHistogramVec(opts, c.Labels)
		} else {
			collector = prometheus.NewHistogram(opts)
		}
	case Gauge:
		opts := prometheus.GaugeOpts{
			Name:      c.Name,
			Namespace: c.Namespace,
			Subsystem: c.Subsystem,
			Help:      c.Help,
		}

		if len(c.Labels) != 0 {
			collector = prometheus.NewGaugeVec(opts, c.Labels)
		} else {
			collector = prometheus.NewGauge(opts)
		}
	case Counter:
		opts := prometheus.CounterOpts{
			Name:      c.Name,
			Namespace: c.Namespace,
			Subsystem: c.Subsystem,
			Help:      c.Help,
		}

		if len(c.Labels) != 0 {
			collector = prometheus.NewCounterVec(opts, c.Labels)
		} else {
			collector = prometheus.NewCounter(opts)
		}
	case Summary:
		opts := prometheus.SummaryOpts{
			Name:       c.Name,
			Namespace:  c.Namespace,
			Subsystem:  c.Subsystem,
			Help:       c.Help,
			Objectives: c.Objectives,
		}

		if len(c.Labels) != 0 {
			collector = prometheus.NewSummaryVec(opts, c.Labels)
		} else {
			collector = prometheus.NewSummary(opts)
		}

	default:
		return fmt.Errorf("unknown collector type `%s`", c.Type)

	}

	// add collector to sync.Map
	rpc.svc.collectors.Store(c.Name, collector)
	// that method might panic, we handle it by recover
	rpc.svc.MustRegister(collector)

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

	switch c := c.(type) {
	case prometheus.Gauge:
		c.Set(m.Value)

	case *prometheus.GaugeVec:
		if len(m.Labels) == 0 {
			return fmt.Errorf("required labels for collector `%s`", m.Name)
		}

		c.WithLabelValues(m.Labels...).Set(m.Value)

	default:
		return fmt.Errorf("collector `%s` does not support method `Set`", m.Name)
	}

	// RPC, set ok to true as return value. Need by rpc.Call reply argument
	*ok = true
	return nil
}
