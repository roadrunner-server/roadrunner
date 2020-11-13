package tests

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spiral/roadrunner/v2/plugins/config"
)

// Gauge //////////////
type Plugin1 struct {
	config config.Configurer
}

func (p1 *Plugin1) Init(cfg config.Configurer) error {
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

func (p1 *Plugin1) MetricsCollector() prometheus.Collector {
	collector := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "my_gauge",
		Help: "My gauge value",
	})

	collector.Set(100)
	return collector
}

// //////////////////////////////////////////////////////////////
type Plugin3 struct {
	config config.Configurer
}

func (p *Plugin3) Init(cfg config.Configurer) error {
	p.config = cfg
	return nil
}

func (p *Plugin3) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin3) Stop() error {
	return nil
}

func (p *Plugin3) Name() string {
	return "metrics_test.plugin3"
}

func (p *Plugin3) MetricsCollector() prometheus.Collector {
	var (
		cpuTemp = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cpu_temperature_celsius",
			Help: "Current temperature of the CPU.",
		})
	)
	return cpuTemp
}

type Plugin4 struct {
	config config.Configurer
}

func (p *Plugin4) Init(cfg config.Configurer) error {
	p.config = cfg
	return nil
}

func (p *Plugin4) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin4) Stop() error {
	return nil
}

func (p *Plugin4) Name() string {
	return "metrics_test.plugin4"
}

func (p *Plugin4) MetricsCollector() prometheus.Collector {
	var (
		cpuTemp = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cpu_temperature_celsius",
			Help: "Current temperature of the CPU.",
		})
	)
	return cpuTemp
}
