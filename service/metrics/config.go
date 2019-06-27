package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spiral/roadrunner/service"
)

type Config struct {
	// Address to listen
	Address string

	// Collect define application specific metrics.
	Collect map[string]Collector
}

// Collector describes single application specific metric.
type Collector struct {
	// Namespace of the metric.
	Namespace string

	// Subsystem of the metric.
	Subsystem string

	// Collector type (histogram, gauge, counter, summary).
	Type string

	// Help of collector.
	Help string

	// Labels for vectorized metrics.
	Labels []string

	// Buckets for histogram metric.
	Buckets []float64
}

// Hydrate configuration.
func (c *Config) Hydrate(cfg service.Config) error {
	return cfg.Unmarshal(c)
}

// register application specific metrics.
func (c *Config) getCollectors() (map[string]prometheus.Collector, error) {
	if c.Collect == nil {
		return nil, nil
	}

	collectors := make(map[string]prometheus.Collector)

	for name, m := range c.Collect {
		var collector prometheus.Collector
		switch m.Type {
		case "histogram":
			opts := prometheus.HistogramOpts{
				Name:      name,
				Namespace: m.Namespace,
				Subsystem: m.Subsystem,
				Help:      m.Help,
				Buckets:   m.Buckets,
			}

			if len(m.Labels) != 0 {
				collector = prometheus.NewHistogramVec(opts, m.Labels)
			} else {
				collector = prometheus.NewHistogram(opts)
			}
		case "gauge":
			opts := prometheus.GaugeOpts{
				Name:      name,
				Namespace: m.Namespace,
				Subsystem: m.Subsystem,
				Help:      m.Help,
			}

			if len(m.Labels) != 0 {
				collector = prometheus.NewGaugeVec(opts, m.Labels)
			} else {
				collector = prometheus.NewGauge(opts)
			}
		case "counter":
			opts := prometheus.CounterOpts{
				Name:      name,
				Namespace: m.Namespace,
				Subsystem: m.Subsystem,
				Help:      m.Help,
			}

			if len(m.Labels) != 0 {
				collector = prometheus.NewCounterVec(opts, m.Labels)
			} else {
				collector = prometheus.NewCounter(opts)
			}
		case "summary":
			opts := prometheus.SummaryOpts{
				Name:      name,
				Namespace: m.Namespace,
				Subsystem: m.Subsystem,
				Help:      m.Help,
			}

			if len(m.Labels) != 0 {
				collector = prometheus.NewSummaryVec(opts, m.Labels)
			} else {
				collector = prometheus.NewSummary(opts)
			}
		default:
			return nil, fmt.Errorf("invalid metric type `%s` for `%s`", m.Type, name)
		}

		collectors[name] = collector
	}

	return collectors, nil
}
