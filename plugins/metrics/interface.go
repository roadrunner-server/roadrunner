package metrics

import "github.com/prometheus/client_golang/prometheus"

type StatProvider interface {
	MetricsCollector() []prometheus.Collector
}
