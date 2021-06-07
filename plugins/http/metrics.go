package http

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	handler "github.com/spiral/roadrunner/v2/pkg/worker_handler"
)

func (p *Plugin) MetricsCollector() []prometheus.Collector {
	// p - implements Exporter interface (workers)
	// other - request duration and count
	return []prometheus.Collector{p, p.requestsExporter.requestDuration, p.requestsExporter.requestCounter}
}

func (p *Plugin) metricsCallback(event interface{}) {
	switch e := event.(type) {
	case handler.ResponseEvent:
		p.requestsExporter.requestCounter.With(prometheus.Labels{
			"status": strconv.Itoa(e.Response.Status),
		}).Inc()

		p.requestsExporter.requestDuration.With(prometheus.Labels{
			"status": strconv.Itoa(e.Response.Status),
		}).Observe(e.Elapsed().Seconds())
	case handler.ErrorEvent:
		p.requestsExporter.requestCounter.With(prometheus.Labels{
			"status": "500",
		}).Inc()

		p.requestsExporter.requestDuration.With(prometheus.Labels{
			"status": "500",
		}).Observe(e.Elapsed().Seconds())
	}
}

type workersExporter struct {
	wm            *prometheus.Desc
	workersMemory uint64
}

func newWorkersExporter() *workersExporter {
	return &workersExporter{
		wm:            prometheus.NewDesc("rr_http_workers_memory_bytes", "Memory usage by HTTP workers.", nil, nil),
		workersMemory: 0,
	}
}

func (p *Plugin) Describe(d chan<- *prometheus.Desc) {
	// send description
	d <- p.workersExporter.wm
}

func (p *Plugin) Collect(ch chan<- prometheus.Metric) {
	// get the copy of the processes
	workers := p.Workers()

	// cumulative RSS memory in bytes
	var cum uint64

	// collect the memory
	for i := 0; i < len(workers); i++ {
		cum += workers[i].MemoryUsage
	}

	// send the values to the prometheus
	ch <- prometheus.MustNewConstMetric(p.workersExporter.wm, prometheus.GaugeValue, float64(cum))
}

type requestsExporter struct {
	requestCounter  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func newRequestsExporter() *requestsExporter {
	return &requestsExporter{
		requestCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rr_http_request_total",
				Help: "Total number of handled http requests after server restart.",
			},
			[]string{"status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "rr_http_request_duration_seconds",
				Help: "HTTP request duration.",
			},
			[]string{"status"},
		),
	}
}
