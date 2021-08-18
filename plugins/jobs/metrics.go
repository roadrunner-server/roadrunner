package jobs

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/plugins/informer"
)

func (p *Plugin) MetricsCollector() []prometheus.Collector {
	// p - implements Exporter interface (workers)
	// other - request duration and count
	return []prometheus.Collector{p.statsExporter}
}

const (
	namespace = "rr_jobs"
)

type statsExporter struct {
	workers       informer.Informer
	workersMemory uint64
	jobsOk        uint64
	pushOk        uint64
	jobsErr       uint64
	pushErr       uint64
}

var (
	worker  = prometheus.NewDesc("workers_memory_bytes", "Memory usage by JOBS workers.", nil, nil)
	pushOk  = prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "push_ok"), "Number of job push.", nil, nil)
	pushErr = prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "push_err"), "Number of jobs push which was failed.", nil, nil)
	jobsErr = prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "jobs_err"), "Number of jobs error while processing in the worker.", nil, nil)
	jobsOk  = prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "jobs_ok"), "Number of successfully processed jobs.", nil, nil)
)

func newStatsExporter(stats informer.Informer) *statsExporter {
	return &statsExporter{
		workers:       stats,
		workersMemory: 0,
		jobsOk:        0,
		pushOk:        0,
		jobsErr:       0,
		pushErr:       0,
	}
}

func (se *statsExporter) metricsCallback(event interface{}) {
	if jev, ok := event.(events.JobEvent); ok {
		switch jev.Event { //nolint:exhaustive
		case events.EventJobOK:
			atomic.AddUint64(&se.jobsOk, 1)
		case events.EventPushOK:
			atomic.AddUint64(&se.pushOk, 1)
		case events.EventPushError:
			atomic.AddUint64(&se.pushErr, 1)
		case events.EventJobError:
			atomic.AddUint64(&se.jobsErr, 1)
		}
	}
}

func (se *statsExporter) Describe(d chan<- *prometheus.Desc) {
	// send description
	d <- pushErr
	d <- pushOk
	d <- jobsErr
	d <- jobsOk
}

func (se *statsExporter) Collect(ch chan<- prometheus.Metric) {
	// get the copy of the processes
	workers := se.workers.Workers()

	// cumulative RSS memory in bytes
	var cum uint64

	// collect the memory
	for i := 0; i < len(workers); i++ {
		cum += workers[i].MemoryUsage
	}

	// send the values to the prometheus
	ch <- prometheus.MustNewConstMetric(worker, prometheus.GaugeValue, float64(cum))
	// send the values to the prometheus
	ch <- prometheus.MustNewConstMetric(jobsOk, prometheus.GaugeValue, float64(atomic.LoadUint64(&se.jobsOk)))
	ch <- prometheus.MustNewConstMetric(jobsErr, prometheus.GaugeValue, float64(atomic.LoadUint64(&se.jobsErr)))
	ch <- prometheus.MustNewConstMetric(pushOk, prometheus.GaugeValue, float64(atomic.LoadUint64(&se.pushOk)))
	ch <- prometheus.MustNewConstMetric(pushErr, prometheus.GaugeValue, float64(atomic.LoadUint64(&se.pushErr)))
}
