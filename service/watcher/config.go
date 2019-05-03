package watcher

import (
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service"
	"time"
)

// Configures set of Services.
type Config struct {
	// Interval defines the update duration for underlying watchers, default 1s.
	Interval time.Duration

	// Services declares list of services to be watched.
	Services map[string]*watcherConfig
}

// Hydrate must populate Config values using given Config source. Must return error if Config is not valid.
func (c *Config) Hydrate(cfg service.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	// Always use second based definition for time durations
	if c.Interval < time.Microsecond {
		c.Interval = time.Second * time.Duration(c.Interval.Nanoseconds())
	}

	return nil
}

// InitDefaults sets missing values to their default values.
func (c *Config) InitDefaults() error {
	c.Interval = time.Second

	return nil
}

// Watchers returns list of defined Services
func (c *Config) Watchers(l listener) (watchers map[string]roadrunner.Watcher) {
	watchers = make(map[string]roadrunner.Watcher)

	for name, cfg := range c.Services {
		watchers[name] = &watcher{lsn: l, tick: c.Interval, cfg: cfg}
	}

	return watchers
}
