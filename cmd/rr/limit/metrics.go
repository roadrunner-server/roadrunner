package limit

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	rr "github.com/spiral/roadrunner/cmd/rr/cmd"
	rrlimit "github.com/spiral/roadrunner/service/limit"
	"github.com/spiral/roadrunner/service/metrics"
)

func init() {
	cobra.OnInitialize(func() {
		svc, _ := rr.Container.Get(metrics.ID)
		mtr, ok := svc.(*metrics.Service)
		if !ok || !mtr.Enabled() {
			return
		}

		ht, _ := rr.Container.Get(rrlimit.ID)
		if ht, ok := ht.(*rrlimit.Service); ok {
			collector := newCollector()

			// register metrics
			mtr.MustRegister(collector.maxMemory)

			// collect events
			ht.AddListener(collector.listener)
		}
	})
}

// listener provide debug callback for system events. With colors!
type metricCollector struct {
	maxMemory        prometheus.Counter
	maxExecutionTime prometheus.Counter
}

func newCollector() *metricCollector {
	return &metricCollector{
		maxMemory: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "rr_limit_max_memory",
				Help: "Total number of workers that was killed because they reached max memory limit.",
			},
		),
		maxExecutionTime: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "rr_limit_max_execution_time",
				Help: "Total number of workers that was killed because they reached max execution time limit.",
			},
		),
	}
}

// listener listens to http events and generates nice looking output.
func (c *metricCollector) listener(event int, ctx interface{}) {
	switch event {
	case rrlimit.EventMaxMemory:
		c.maxMemory.Inc()
	case rrlimit.EventExecTTL:
		c.maxExecutionTime.Inc()
	}
}
