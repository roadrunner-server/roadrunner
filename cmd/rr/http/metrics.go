package http

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	rr "github.com/spiral/roadrunner/cmd/rr/cmd"
	rrhttp "github.com/spiral/roadrunner/service/http"
	"github.com/spiral/roadrunner/service/metrics"
	"github.com/spiral/roadrunner/util"
	"strconv"
	"time"
)

func init() {
	cobra.OnInitialize(func() {
		svc, _ := rr.Container.Get(metrics.ID)
		mtr, ok := svc.(*metrics.Service)
		if !ok || !mtr.Enabled() {
			return
		}

		ht, _ := rr.Container.Get(rrhttp.ID)
		if ht, ok := ht.(*rrhttp.Service); ok {
			collector := newCollector()

			// register metrics
			mtr.MustRegister(collector.requestCounter)
			mtr.MustRegister(collector.requestDuration)
			mtr.MustRegister(collector.workersMemory)

			// collect events
			ht.AddListener(collector.listener)

			// update memory usage every 10 seconds
			go collector.collectMemory(ht, time.Second*10)
		}
	})
}

// listener provide debug callback for system events. With colors!
type metricCollector struct {
	requestCounter  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	workersMemory   prometheus.Gauge
}

func newCollector() *metricCollector {
	return &metricCollector{
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
		workersMemory: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "rr_http_workers_memory_bytes",
				Help: "Memory usage by HTTP workers.",
			},
		),
	}
}

// listener listens to http events and generates nice looking output.
func (c *metricCollector) listener(event int, ctx interface{}) {
	// http events
	switch event {
	case rrhttp.EventResponse:
		e := ctx.(*rrhttp.ResponseEvent)

		c.requestCounter.With(prometheus.Labels{
			"status": strconv.Itoa(e.Response.Status),
		}).Inc()

		c.requestDuration.With(prometheus.Labels{
			"status": strconv.Itoa(e.Response.Status),
		}).Observe(e.Elapsed().Seconds())

	case rrhttp.EventError:
		e := ctx.(*rrhttp.ErrorEvent)

		c.requestCounter.With(prometheus.Labels{
			"status": "500",
		}).Inc()

		c.requestDuration.With(prometheus.Labels{
			"status": "500",
		}).Observe(e.Elapsed().Seconds())
	}
}

// collect memory usage by server workers
func (c *metricCollector) collectMemory(service *rrhttp.Service, tick time.Duration) {
	started := false
	for {
		server := service.Server()
		if server == nil && started {
			// stopped
			return
		}

		started = true

		if workers, err := util.ServerState(server); err == nil {
			sum := 0.0
			for _, w := range workers {
				sum = sum + float64(w.MemoryUsage)
			}

			c.workersMemory.Set(sum)
		}

		time.Sleep(tick)
	}
}
