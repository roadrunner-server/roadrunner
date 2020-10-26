package roadrunner

import (
	"context"
	"time"

	"github.com/spiral/roadrunner/v2/util"
)

const MB = 1024 * 1024

type SupervisedPool interface {
	Pool

	// ExecWithContext provides the ability to execute with time deadline. Attention, worker will be destroyed if context
	// deadline reached.
	ExecWithContext(ctx context.Context, rqs Payload) (Payload, error)
}

type supervisedPool struct {
	cfg    SupervisorConfig
	events *util.EventHandler
	pool   Pool
	stopCh chan struct{}
}

func newPoolWatcher(pool *StaticPool, events *util.EventHandler, cfg SupervisorConfig) *supervisedPool {
	return &supervisedPool{
		cfg:    cfg,
		events: events,
		pool:   pool,
		stopCh: make(chan struct{}),
	}
}

func (sp *supervisedPool) Start() {
	go func() {
		watchTout := time.NewTicker(sp.cfg.WatchTick)
		for {
			select {
			case <-sp.stopCh:
				watchTout.Stop()
				return
			// stop here
			case <-watchTout.C:
				sp.control()
			}
		}
	}()
}

func (sp *supervisedPool) Stop() {
	sp.stopCh <- struct{}{}
}

func (sp *supervisedPool) control() {
	now := time.Now()
	ctx := context.TODO()

	// THIS IS A COPY OF WORKERS
	workers := sp.pool.Workers()

	for i := 0; i < len(workers); i++ {
		if workers[i].State().Value() == StateInvalid {
			continue
		}

		s, err := WorkerProcessState(workers[i])
		if err != nil {
			// worker not longer valid for supervision
			continue
		}

		if sp.cfg.TTL != 0 && now.Sub(workers[i].Created()).Seconds() >= float64(sp.cfg.TTL) {
			err = sp.pool.RemoveWorker(ctx, workers[i])
			if err != nil {
				sp.events.Push(PoolEvent{Event: EventSupervisorError, Payload: err})
				return
			} else {
				sp.events.Push(PoolEvent{Event: EventTTL, Payload: workers[i]})
			}

			continue
		}

		if sp.cfg.MaxWorkerMemory != 0 && s.MemoryUsage >= sp.cfg.MaxWorkerMemory*MB {
			err = sp.pool.RemoveWorker(ctx, workers[i])
			if err != nil {
				sp.events.Push(PoolEvent{Event: EventSupervisorError, Payload: err})
				return
			} else {
				sp.events.Push(PoolEvent{Event: EventTTL, Payload: workers[i]})
			}

			continue
		}

		// firs we check maxWorker idle
		if sp.cfg.IdleTTL != 0 {
			// then check for the worker state
			if workers[i].State().Value() != StateReady {
				continue
			}

			/*
				Calculate idle time
				If worker in the StateReady, we read it LastUsed timestamp as UnixNano uint64
				2. For example maxWorkerIdle is equal to 5sec, then, if (time.Now - LastUsed) > maxWorkerIdle
				we are guessing that worker overlap idle time and has to be killed
			*/

			// get last used unix nano
			lu := workers[i].State().LastUsed()

			// convert last used to unixNano and sub time.now
			res := int64(lu) - now.UnixNano()

			// maxWorkerIdle more than diff between now and last used
			if sp.cfg.IdleTTL-res <= 0 {
				err = sp.pool.RemoveWorker(ctx, workers[i])
				if err != nil {
					sp.events.Push(PoolEvent{Event: EventSupervisorError, Payload: err})
					return
				} else {
					sp.events.Push(PoolEvent{Event: EventIdleTTL, Payload: workers[i]})
				}
			}
		}
	}
}
