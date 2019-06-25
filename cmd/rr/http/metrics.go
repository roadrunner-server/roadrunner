package http

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	rr "github.com/spiral/roadrunner/cmd/rr/cmd"
	rrhttp "github.com/spiral/roadrunner/service/http"
	"github.com/spiral/roadrunner/service/metrics"
	"strconv"
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

			// collect events
			ht.AddListener(collector.listener)
		}
	})
}

// listener provide debug callback for system events. With colors!
type metricCollector struct {
	requestCounter  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func newCollector() *metricCollector {
	return &metricCollector{
		requestCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rr_http_total",
				Help: "Total number of handled http requests after server restart.",
			},
			[]string{"status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "rr_http_request_duration",
				Help:    "HTTP request duration.",
				Buckets: []float64{0.25, 0.5, 1, 10, 20, 60},
			},
			[]string{"status"},
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
