package pool

import (
	"context"
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/payload"
	"github.com/spiral/roadrunner/v2/pkg/process"
	"github.com/spiral/roadrunner/v2/pkg/worker"
)

const MB = 1024 * 1024

// NSEC_IN_SEC nanoseconds in second
const NSEC_IN_SEC int64 = 1000000000 //nolint:stylecheck

type Supervised interface {
	Pool
	// Start used to start watching process for all pool workers
	Start()
}

type supervised struct {
	cfg    *SupervisorConfig
	events events.Handler
	pool   Pool
	stopCh chan struct{}
	mu     *sync.RWMutex
}

func supervisorWrapper(pool Pool, events events.Handler, cfg *SupervisorConfig) Supervised {
	sp := &supervised{
		cfg:    cfg,
		events: events,
		pool:   pool,
		mu:     &sync.RWMutex{},
		stopCh: make(chan struct{}),
	}

	return sp
}

func (sp *supervised) execWithTTL(_ context.Context, _ payload.Payload) (payload.Payload, error) {
	panic("used to satisfy pool interface")
}

func (sp *supervised) Exec(rqs payload.Payload) (payload.Payload, error) {
	const op = errors.Op("supervised_exec_with_context")
	if sp.cfg.ExecTTL == 0 {
		return sp.pool.Exec(rqs)
	}

	ctx, cancel := context.WithTimeout(context.Background(), sp.cfg.ExecTTL)
	defer cancel()

	res, err := sp.pool.execWithTTL(ctx, rqs)
	if err != nil {
		return payload.Payload{}, errors.E(op, err)
	}

	return res, nil
}

func (sp *supervised) GetConfig() interface{} {
	return sp.pool.GetConfig()
}

func (sp *supervised) Workers() (workers []worker.BaseProcess) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	return sp.pool.Workers()
}

func (sp *supervised) RemoveWorker(worker worker.BaseProcess) error {
	return sp.pool.RemoveWorker(worker)
}

func (sp *supervised) Destroy(ctx context.Context) {
	sp.pool.Destroy(ctx)
}

func (sp *supervised) Start() {
	go func() {
		watchTout := time.NewTicker(sp.cfg.WatchTick)
		for {
			select {
			case <-sp.stopCh:
				watchTout.Stop()
				return
			// stop here
			case <-watchTout.C:
				sp.mu.Lock()
				sp.control()
				sp.mu.Unlock()
			}
		}
	}()
}

func (sp *supervised) Stop() {
	sp.stopCh <- struct{}{}
}

func (sp *supervised) control() { //nolint:gocognit
	now := time.Now()

	// MIGHT BE OUTDATED
	// It's a copy of the Workers pointers
	workers := sp.pool.Workers()

	for i := 0; i < len(workers); i++ {
		// if worker not in the Ready OR working state
		// skip such worker
		switch workers[i].State().Value() {
		case
			worker.StateInvalid,
			worker.StateErrored,
			worker.StateDestroyed,
			worker.StateInactive,
			worker.StateStopped,
			worker.StateStopping:
			continue
		}

		s, err := process.WorkerProcessState(workers[i])
		if err != nil {
			// worker not longer valid for supervision
			continue
		}

		if sp.cfg.TTL != 0 && now.Sub(workers[i].Created()).Seconds() >= sp.cfg.TTL.Seconds() {
			// SOFT termination. DO NOT STOP active workers
			if workers[i].State().Value() != worker.StateWorking {
				workers[i].State().Set(worker.StateInvalid)
				_ = workers[i].Stop()
			}
			sp.events.Push(events.PoolEvent{Event: events.EventTTL, Payload: workers[i]})
			continue
		}

		if sp.cfg.MaxWorkerMemory != 0 && s.MemoryUsage >= sp.cfg.MaxWorkerMemory*MB {
			// SOFT termination. DO NOT STOP active workers
			if workers[i].State().Value() != worker.StateWorking {
				workers[i].State().Set(worker.StateInvalid)
				_ = workers[i].Stop()
			}

			// mark it as invalid, worker likely in the StateWorking, so, it will be killed after work will be done
			workers[i].State().Set(worker.StateInvalid)
			sp.events.Push(events.PoolEvent{Event: events.EventMaxMemory, Payload: workers[i]})
			continue
		}

		// firs we check maxWorker idle
		if sp.cfg.IdleTTL != 0 {
			// then check for the worker state
			if workers[i].State().Value() != worker.StateReady {
				continue
			}

			/*
				Calculate idle time
				If worker in the StateReady, we read it LastUsed timestamp as UnixNano uint64
				2. For example maxWorkerIdle is equal to 5sec, then, if (time.Now - LastUsed) > maxWorkerIdle
				we are guessing that worker overlap idle time and has to be killed
			*/

			// 1610530005534416045 lu
			// lu - now = -7811150814 - nanoseconds
			// 7.8 seconds
			// get last used unix nano
			lu := workers[i].State().LastUsed()
			// worker not used, skip
			if lu == 0 {
				continue
			}

			// convert last used to unixNano and sub time.now to seconds
			// negative number, because lu always in the past, except for the `back to the future` :)
			res := ((int64(lu) - now.UnixNano()) / NSEC_IN_SEC) * -1

			// maxWorkerIdle more than diff between now and last used
			// for example:
			// After exec worker goes to the rest
			// And resting for the 5 seconds
			// IdleTTL is 1 second.
			// After the control check, res will be 5, idle is 1
			// 5 - 1 = 4, more than 0, YOU ARE FIRED (removed). Done.
			if int64(sp.cfg.IdleTTL.Seconds())-res <= 0 {
				if workers[i].State().Value() != worker.StateWorking {
					workers[i].State().Set(worker.StateInvalid)
					_ = workers[i].Stop()
				}

				workers[i].State().Set(worker.StateInvalid)
				sp.events.Push(events.PoolEvent{Event: events.EventIdleTTL, Payload: workers[i]})
			}
		}
	}
}
