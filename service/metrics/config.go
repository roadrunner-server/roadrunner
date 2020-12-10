package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spiral/roadrunner/service"
)

// Config configures metrics service.
type Config struct {
	// Address to listen
	Address string

	// Collect define application specific metrics.
	Collect map[string]Collector
}

type NamedCollector struct {
	// Name of the collector
	Name string `json:"name"`

	// Collector structure
	Collector `json:"collector"`
}

// CollectorType represents prometheus collector types
type CollectorType string

const (
	// Histogram type
	Histogram CollectorType = "histogram"

	// Gauge type
	Gauge CollectorType = "gauge"

	// Counter type
	Counter CollectorType = "counter"

	// Summary type
	Summary CollectorType = "summary"
)

// Collector describes single application specific metric.
type Collector struct {
	// Namespace of the metric.
	Namespace string `json:"namespace"`
	// Subsystem of the metric.
	Subsystem string `json:"subsystem"`
	// Collector type (histogram, gauge, counter, summary).
	Type CollectorType `json:"type"`
	// Help of collector.
	Help string `json:"help"`
	// Labels for vectorized metrics.
	Labels []string `json:"labels"`
	// Buckets for histogram metric.
	Buckets []float64 `json:"buckets"`
	// Objectives for the summary opts
	Objectives map[float64]float64 `json:"objectives"`
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
		case Histogram:
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
		case Gauge:
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
		case Counter:
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
		case Summary:
			opts := prometheus.SummaryOpts{
				Name:       name,
				Namespace:  m.Namespace,
				Subsystem:  m.Subsystem,
				Help:       m.Help,
				Objectives: m.Objectives,
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
