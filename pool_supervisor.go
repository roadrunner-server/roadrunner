package roadrunner

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const MB = 1024 * 1024

type Supervisor interface {
	Attach(pool Pool)
	StartWatching() error
	StopWatching()
	Detach()
}

type staticPoolSupervisor struct {
	// maxWorkerMemory in MB
	maxWorkerMemory uint64
	// maxPoolMemory in MB
	maxPoolMemory uint64
	// maxWorkerTTL in seconds
	maxWorkerTTL uint64
	// maxWorkerIdle in seconds
	maxWorkerIdle uint64

	// watchTimeout in seconds
	watchTimeout uint64
	stopCh       chan struct{}

	pool Pool
}

/*
The arguments are:
maxWorkerMemory - maximum memory allowed for a single worker
maxPoolMemory - maximum pool memory allowed for a pool of a workers
maxTtl - maximum ttl for the worker after which it will be killed and replaced
maxIdle - maximum time to live for the worker in Ready state
watchTimeout - time between watching for the workers/pool status
*/
// TODO might be just wrap the pool and return ControlledPool with included Pool interface
func NewStaticPoolSupervisor(maxWorkerMemory, maxPoolMemory, maxTtl, maxIdle, watchTimeout uint64) Supervisor {
	if maxWorkerMemory == 0 {
		// just set to a big number, 5GB
		maxPoolMemory = 5000 * MB
	}
	if watchTimeout == 0 {
		watchTimeout = 60
	}
	return &staticPoolSupervisor{
		maxWorkerMemory: maxWorkerMemory,
		maxPoolMemory:   maxPoolMemory,
		maxWorkerTTL:    maxTtl,
		maxWorkerIdle:   maxIdle,
		stopCh:          make(chan struct{}),
	}
}

func (sps *staticPoolSupervisor) Attach(pool Pool) {
	sps.pool = pool
}

func (sps *staticPoolSupervisor) StartWatching() error {
	go func() {
		watchTout := time.NewTicker(time.Second * time.Duration(sps.watchTimeout))
		for {
			select {
			case <-sps.stopCh:
				watchTout.Stop()
				return
			// stop here
			case <-watchTout.C:
				err := sps.control()
				if err != nil {
					sps.pool.Events() <- PoolEvent{Payload: err}
				}
			}
		}
	}()
	return nil
}

func (sps *staticPoolSupervisor) StopWatching() {
	sps.stopCh <- struct{}{}
}

func (sps *staticPoolSupervisor) Detach() {

}

func (sps *staticPoolSupervisor) control() error {
	if sps.pool == nil {
		return errors.New("pool should be attached")
	}
	now := time.Now()
	ctx := context.TODO()

	// THIS IS A COPY OF WORKERS
	workers := sps.pool.Workers(ctx)
	var totalUsedMemory uint64

	for i := 0; i < len(workers); i++ {
		if workers[i].State().Value() == StateInvalid {
			continue
		}

		s, err := WorkerProcessState(workers[i])
		if err != nil {
			panic(err)
			// push to pool events??
		}

		if sps.maxWorkerTTL != 0 && now.Sub(workers[i].Created()).Seconds() >= float64(sps.maxWorkerTTL) {
			err = sps.pool.RemoveWorker(ctx, workers[i])
			if err != nil {
				return err
			}

			// after remove worker we should exclude it from further analysis
			workers = append(workers[:i], workers[i+1:]...)
		}

		if sps.maxWorkerMemory != 0 && s.MemoryUsage >= sps.maxWorkerMemory*MB {
			// TODO events
			sps.pool.Events() <- PoolEvent{Payload: fmt.Errorf("max allowed memory reached (%vMB)", sps.maxWorkerMemory)}
			err = sps.pool.RemoveWorker(ctx, workers[i])
			if err != nil {
				return err
			}
			workers = append(workers[:i], workers[i+1:]...)
			continue
		}

		// firs we check maxWorker idle
		if sps.maxWorkerIdle != 0 {
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
			if int64(sps.maxWorkerIdle)-res <= 0 {
				sps.pool.Events() <- PoolEvent{Payload: fmt.Errorf("max allowed worker idle time elapsed. actual idle time: %v, max idle time: %v", sps.maxWorkerIdle, res)}
				err = sps.pool.RemoveWorker(ctx, workers[i])
				if err != nil {
					return err
				}
				workers = append(workers[:i], workers[i+1:]...)
			}
		}

		// the very last step is to calculate pool memory usage (except excluded workers)
		totalUsedMemory += s.MemoryUsage
	}

	// if current usage more than max allowed pool memory usage
	if totalUsedMemory > sps.maxPoolMemory {
		// destroy pool
		totalUsedMemory = 0
		sps.pool.Destroy(ctx)
	}

	return nil
}
