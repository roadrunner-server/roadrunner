package tests

import (
	"github.com/prometheus/client_golang/prometheus"
	config2 "github.com/spiral/roadrunner/v2/interfaces/config"
)

// Gauge //////////////
type Plugin1 struct {
	config config2.Configurer
}

func (p1 *Plugin1) Init(cfg config2.Configurer) error {
	p1.config = cfg
	return nil
}

func (p1 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p1 *Plugin1) Stop() error {
	return nil
}

func (p1 *Plugin1) Name() string {
	return "metrics_test.plugin1"
}

func (p1 *Plugin1) MetricsCollector() []prometheus.Collector {
	collector := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "my_gauge",
		Help: "My gauge value",
	})

	collector.Set(100)

	collector2 := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "my_gauge2",
		Help: "My gauge2 value",
	})

	collector2.Set(100)
	return []prometheus.Collector{collector, collector2}
}
