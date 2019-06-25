package metrics

import "github.com/spiral/roadrunner/service"

type Config struct {
	// Address to listen
	Address string

	// Metrics define application specific metrics.
	Metrics map[string]Metric
}

// Metric describes single application specific metric.
type Metric struct {
	Type        string
	Description string
	Labels      []string
}

// Hydrate configuration.
func (c *Config) Hydrate(cfg service.Config) error {
	return cfg.Unmarshal(c)
}
