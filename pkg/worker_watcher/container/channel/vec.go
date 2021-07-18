package channel

import (
	"context"
	"sync/atomic"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/worker"
)

type Vec struct {
	destroy uint64
	workers chan worker.BaseProcess
}

func NewVector(initialNumOfWorkers uint64) *Vec {
	vec := &Vec{
		destroy: 0,
		workers: make(chan worker.BaseProcess, initialNumOfWorkers),
	}

	return vec
}

func (v *Vec) Push(w worker.BaseProcess) {
	v.workers <- w
}

func (v *Vec) Remove(_ int64) {}

func (v *Vec) Pop(ctx context.Context) (worker.BaseProcess, error) {
	/*
		if *addr == old {
			*addr = new
			return true
		}
	*/

	if atomic.CompareAndSwapUint64(&v.destroy, 1, 1) {
		return nil, errors.E(errors.WatcherStopped)
	}

	select {
	case w := <-v.workers:
		return w, nil
	case <-ctx.Done():
		return nil, errors.E(ctx.Err(), errors.NoFreeWorkers)
	}
}

func (v *Vec) Destroy() {
	atomic.StoreUint64(&v.destroy, 1)
}
