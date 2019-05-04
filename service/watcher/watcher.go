package watcher

import (
	"fmt"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/util"
	"time"
)

const (
	// EventMaxMemory caused when worker consumes more memory than allowed.
	EventMaxMemory = iota + 8000

	// EventMaxTTL thrown when worker is removed due TTL being reached. Context is rr.WorkerError
	EventMaxTTL

	// EventMaxIdleTTL triggered when worker spends too much time at rest.
	EventMaxIdleTTL

	// EventMaxIdleTTL triggered when worker spends too much time doing the task (max_execution_time).
	EventMaxExecTTL
)

// handles watcher events
type listener func(event int, ctx interface{})

// defines the watcher behaviour
type watcherConfig struct {
	// MaxMemory defines maximum amount of memory allowed for worker. In megabytes.
	MaxMemory uint64

	// TTL defines maximum time worker is allowed to live.
	TTL int64

	// MaxIdleTTL defines maximum duration worker can spend in idle mode.
	MaxIdleTTL int64

	// MaxExecTTL defines maximum lifetime per job.
	MaxExecTTL int64
}

type watcher struct {
	lsn  listener
	tick time.Duration
	cfg  *watcherConfig

	// list of workers which are currently working
	sw *stateWatcher

	stop chan interface{}
}

// watch the pool state
func (wch *watcher) watch(p roadrunner.Pool) {
	now := time.Now()

	for _, w := range p.Workers() {
		if w.State().Value() == roadrunner.StateInvalid {
			// skip duplicate assessment
			continue
		}

		s, err := util.WorkerState(w)
		if err != nil {
			continue
		}

		if wch.cfg.TTL != 0 && now.Sub(w.Created).Seconds() >= float64(wch.cfg.TTL) {
			err := fmt.Errorf("max TTL reached (%vs)", wch.cfg.TTL)
			if p.Remove(w, err) {
				wch.report(EventMaxTTL, w, err)
			}
			continue
		}

		if wch.cfg.MaxMemory != 0 && s.MemoryUsage >= wch.cfg.MaxMemory*1024*1024 {
			err := fmt.Errorf("max allowed memory reached (%vMB)", wch.cfg.MaxMemory)
			if p.Remove(w, err) {
				wch.report(EventMaxMemory, w, err)
			}
			continue
		}

		// watch the worker state changes
		wch.sw.push(w)
	}

	wch.sw.sync(now)

	if wch.cfg.MaxExecTTL != 0 {
		for _, w := range wch.sw.find(
			roadrunner.StateWorking,
			now.Add(-time.Second*time.Duration(wch.cfg.MaxExecTTL)),
		) {
			eID := w.State().NumExecs()
			err := fmt.Errorf("max exec time reached (%vs)", wch.cfg.MaxExecTTL)

			if p.Remove(w, err) {
				// make sure worker still on initial request
				if w.State().NumExecs() == eID {
					go w.Kill()
					wch.report(EventMaxExecTTL, w, err)
				}
			}
		}
	}

	// locale workers which are in idle mode for too long
	if wch.cfg.MaxIdleTTL != 0 {
		for _, w := range wch.sw.find(
			roadrunner.StateReady,
			now.Add(-time.Second*time.Duration(wch.cfg.MaxIdleTTL)),
		) {
			err := fmt.Errorf("max idle time reached (%vs)", wch.cfg.MaxIdleTTL)
			if p.Remove(w, err) {
				wch.report(EventMaxIdleTTL, w, err)
			}
		}
	}
}

// throw watcher event
func (wch *watcher) report(event int, worker *roadrunner.Worker, caused error) {
	if wch.lsn != nil {
		wch.lsn(event, roadrunner.WorkerError{Worker: worker, Caused: caused})
	}
}

// Attach watcher to the pool
func (wch *watcher) Attach(pool roadrunner.Pool) roadrunner.Watcher {
	wp := &watcher{
		tick: wch.tick,
		lsn:  wch.lsn,
		cfg:  wch.cfg,
		sw:   newStateWatcher(),
		stop: make(chan interface{}),
	}

	go func(wp *watcher, pool roadrunner.Pool) {
		ticker := time.NewTicker(wp.tick)
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
func (wch *watcher) Detach() {
	close(wch.stop)
}
