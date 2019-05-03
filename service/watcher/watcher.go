package watcher

import (
	"fmt"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/util"
	"time"
)

const (
	// EventMaxTTL thrown when worker is removed due MaxTTL being reached. Context is roadrunner.WorkerError
	EventMaxTTL = iota + 8000

	// EventMaxMemory caused when worker consumes more memory than allowed.
	EventMaxMemory
)

// handles watcher events
type listener func(event int, ctx interface{})

// defines the watcher behaviour
type watcherConfig struct {
	// MaxTTL defines maximum time worker is allowed to live.
	MaxTTL time.Duration

	// MaxMemory defines maximum amount of memory allowed for worker. In megabytes.
	MaxMemory uint64
}

// Normalize watcher config and upscale the durations.
func (c *watcherConfig) Normalize() error {
	// Always use second based definition for time durations
	if c.MaxTTL < time.Microsecond {
		c.MaxTTL = time.Second * time.Duration(c.MaxTTL.Nanoseconds())
	}

	return nil
}

type watcher struct {
	lsn      listener
	interval time.Duration
	cfg      *watcherConfig
	stop     chan interface{}
}

// watch the pool state
func (watch *watcher) watch(p roadrunner.Pool) {
	now := time.Now()
	for _, w := range p.Workers() {
		if watch.cfg.MaxTTL != 0 && now.Sub(w.Created) >= watch.cfg.MaxTTL {
			err := fmt.Errorf("max TTL reached (%s)", watch.cfg.MaxTTL)
			if p.Remove(w, err) {
				watch.report(EventMaxTTL, w, err)
			}
		}

		state, err := util.WorkerState(w)
		if err != nil {
			continue
		}

		if watch.cfg.MaxMemory != 0 && state.MemoryUsage >= watch.cfg.MaxMemory*1024*1024 {
			err := fmt.Errorf("max allowed memory reached (%vMB)", watch.cfg.MaxMemory)
			if p.Remove(w, err) {
				watch.report(EventMaxMemory, w, err)
			}
		}
	}
}

// throw watcher event
func (watch *watcher) report(event int, worker *roadrunner.Worker, caused error) {
	if watch.lsn != nil {
		watch.lsn(event, roadrunner.WorkerError{Worker: worker, Caused: caused})
	}
}

// Attach watcher to the pool
func (watch *watcher) Attach(pool roadrunner.Pool) roadrunner.Watcher {
	wp := &watcher{
		interval: watch.interval,
		lsn:      watch.lsn,
		cfg:      watch.cfg,
		stop:     make(chan interface{}),
	}

	go func(wp *watcher, pool roadrunner.Pool) {
		ticker := time.NewTicker(wp.interval)
		for {
			select {
			case <-ticker.C:
				wp.watch(pool)
			case <-wp.stop:
				return
			}
		}
	}(wp, pool)

	return wp
}

// Detach watcher from the pool.
func (watch *watcher) Detach() {
	close(watch.stop)
}
