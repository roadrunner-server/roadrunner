package limit

import (
	"fmt"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/util"
	"time"
)

const (
	// EventMaxMemory caused when worker consumes more memory than allowed.
	EventMaxMemory = iota + 8000

	// EventTTL thrown when worker is removed due TTL being reached. Context is rr.WorkerError
	EventTTL

	// EventIdleTTL triggered when worker spends too much time at rest.
	EventIdleTTL

	// EventExecTTL triggered when worker spends too much time doing the task (max_execution_time).
	EventExecTTL
)

// handles controller events
type listener func(event int, ctx interface{})

// defines the controller behaviour
type controllerConfig struct {
	// MaxMemory defines maximum amount of memory allowed for worker. In megabytes.
	MaxMemory uint64

	// TTL defines maximum time worker is allowed to live.
	TTL int64

	// IdleTTL defines maximum duration worker can spend in idle mode.
	IdleTTL int64

	// ExecTTL defines maximum lifetime per job.
	ExecTTL int64
}

type controller struct {
	lsn  listener
	tick time.Duration
	cfg  *controllerConfig

	// list of workers which are currently working
	sw *stateFilter

	stop chan interface{}
}

// control the pool state
func (c *controller) control(p roadrunner.Pool) {
	c.loadWorkers(p)

	now := time.Now()

	if c.cfg.ExecTTL != 0 {
		for _, w := range c.sw.find(
			roadrunner.StateWorking,
			now.Add(-time.Second*time.Duration(c.cfg.ExecTTL)),
		) {
			eID := w.State().NumExecs()
			err := fmt.Errorf("max exec time reached (%vs)", c.cfg.ExecTTL)

			// make sure worker still on initial request
			if p.Remove(w, err) && w.State().NumExecs() == eID {
				go w.Kill()
				c.report(EventExecTTL, w, err)
			}
		}
	}

	// locale workers which are in idle mode for too long
	if c.cfg.IdleTTL != 0 {
		for _, w := range c.sw.find(
			roadrunner.StateReady,
			now.Add(-time.Second*time.Duration(c.cfg.IdleTTL)),
		) {
			err := fmt.Errorf("max idle time reached (%vs)", c.cfg.IdleTTL)
			if p.Remove(w, err) {
				c.report(EventIdleTTL, w, err)
			}
		}
	}
}

func (c *controller) loadWorkers(p roadrunner.Pool) {
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

		if c.cfg.TTL != 0 && now.Sub(w.Created).Seconds() >= float64(c.cfg.TTL) {
			err := fmt.Errorf("max TTL reached (%vs)", c.cfg.TTL)
			if p.Remove(w, err) {
				c.report(EventTTL, w, err)
			}
			continue
		}

		if c.cfg.MaxMemory != 0 && s.MemoryUsage >= c.cfg.MaxMemory*1024*1024 {
			err := fmt.Errorf("max allowed memory reached (%vMB)", c.cfg.MaxMemory)
			if p.Remove(w, err) {
				c.report(EventMaxMemory, w, err)
			}
			continue
		}

		// control the worker state changes
		c.sw.push(w)
	}

	c.sw.sync(now)
}

// throw controller event
func (c *controller) report(event int, worker *roadrunner.Worker, caused error) {
	if c.lsn != nil {
		c.lsn(event, roadrunner.WorkerError{Worker: worker, Caused: caused})
	}
}

// Attach controller to the pool
func (c *controller) Attach(pool roadrunner.Pool) roadrunner.Controller {
	wp := &controller{
		tick: c.tick,
		lsn:  c.lsn,
		cfg:  c.cfg,
		sw:   newStateFilter(),
		stop: make(chan interface{}),
	}

	go func(wp *controller, pool roadrunner.Pool) {
		ticker := time.NewTicker(wp.tick)
		for {
			select {
			case <-ticker.C:
				wp.control(pool)
			case <-wp.stop:
				return
			}
		}
	}(wp, pool)

	return wp
}

// Detach controller from the pool.
func (c *controller) Detach() {
	close(c.stop)
}
