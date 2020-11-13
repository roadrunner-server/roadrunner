package tests

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spiral/roadrunner/v2/plugins/config"
)

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
	var (
		cpuTemp = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cpu_temperature_celsius",
			Help: "Current temperature of the CPU.",
		})
	)
	return cpuTemp
}

type PluginRpc struct {
	srv *Plugin1
}

func (r *PluginRpc) Hello(in string, out *string) error {
	*out = fmt.Sprintf("Hello, username: %s", in)
	return nil
}
