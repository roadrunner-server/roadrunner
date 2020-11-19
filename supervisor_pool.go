package roadrunner

import (
	"context"
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/util"
)

const MB = 1024 * 1024

type SupervisedPool interface {
	Pool
	// Start used to start watching process for all pool workers
	Start()
}

type supervisedPool struct {
	cfg    *SupervisorConfig
	events util.EventsHandler
	pool   Pool
	stopCh chan struct{}
	mu     *sync.RWMutex
}

func newPoolWatcher(pool Pool, events util.EventsHandler, cfg *SupervisorConfig) SupervisedPool {
	sp := &supervisedPool{
		cfg:    cfg,
		events: events,
		pool:   pool,
		mu:     &sync.RWMutex{},
		stopCh: make(chan struct{}),
	}
	return sp
}

type ttlExec struct {
	err error
	p   Payload
}

func (sp *supervisedPool) ExecWithContext(ctx context.Context, rqs Payload) (Payload, error) {
	const op = errors.Op("exec_supervised")
	if sp.cfg.ExecTTL == 0 {
		return sp.pool.Exec(rqs)
	}

	c := make(chan ttlExec, 1)
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(sp.cfg.ExecTTL))
	defer cancel()
	go func() {
		res, err := sp.pool.ExecWithContext(ctx, rqs)
		if err != nil {
			c <- ttlExec{
				err: errors.E(op, err),
				p:   EmptyPayload,
			}
		}

		c <- ttlExec{
			err: nil,
			p:   res,
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return EmptyPayload, errors.E(op, errors.TimeOut, ctx.Err())
		case res := <-c:
			if res.err != nil {
				return EmptyPayload, res.err
			}

			return res.p, nil
		}
	}
}

func (sp *supervisedPool) Exec(p Payload) (Payload, error) {
	const op = errors.Op("supervised exec")
	rsp, err := sp.pool.Exec(p)
	if err != nil {
		return EmptyPayload, errors.E(op, err)
	}
	return rsp, nil
}

func (sp *supervisedPool) AddListener(listener util.EventListener) {
	sp.pool.AddListener(listener)
}

func (sp *supervisedPool) GetConfig() PoolConfig {
	return sp.pool.GetConfig()
}

func (sp *supervisedPool) Workers() (workers []WorkerBase) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	return sp.pool.Workers()
}

func (sp *supervisedPool) RemoveWorker(ctx context.Context, worker WorkerBase) error {
	return sp.pool.RemoveWorker(ctx, worker)
}

func (sp *supervisedPool) Destroy(ctx context.Context) {
	sp.pool.Destroy(ctx)
}

func (sp *supervisedPool) Start() {
	go func() {
		watchTout := time.NewTicker(time.Second * time.Duration(sp.cfg.WatchTick))
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

func (sp *supervisedPool) Stop() {
	sp.stopCh <- struct{}{}
}

func (sp *supervisedPool) control() {
	now := time.Now()
	ctx := context.TODO()
	const op = errors.Op("supervised pool control tick")

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
				sp.events.Push(PoolEvent{Event: EventSupervisorError, Payload: errors.E(op, err)})
				return
			}
			sp.events.Push(PoolEvent{Event: EventTTL, Payload: workers[i]})
			continue
		}

		if sp.cfg.MaxWorkerMemory != 0 && s.MemoryUsage >= sp.cfg.MaxWorkerMemory*MB {
			err = sp.pool.RemoveWorker(ctx, workers[i])
			if err != nil {
				sp.events.Push(PoolEvent{Event: EventSupervisorError, Payload: errors.E(op, err)})
				return
			}
			sp.events.Push(PoolEvent{Event: EventMaxMemory, Payload: workers[i]})

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
			if sp.cfg.IdleTTL-uint64(res) <= 0 {
				err = sp.pool.RemoveWorker(ctx, workers[i])
				if err != nil {
					sp.events.Push(PoolEvent{Event: EventSupervisorError, Payload: errors.E(op, err)})
					return
				}
				sp.events.Push(PoolEvent{Event: EventIdleTTL, Payload: workers[i]})
			}
		}
	}
}
