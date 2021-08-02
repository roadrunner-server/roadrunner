package channel

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/worker"
)

type Vec struct {
	sync.RWMutex
	// destroy signal
	destroy uint64
	// channel with the workers
	workers chan worker.BaseProcess

	len uint64
}

func NewVector(len uint64) *Vec {
	vec := &Vec{
		destroy: 0,
		len:     len,
		workers: make(chan worker.BaseProcess, len),
	}

	return vec
}

// Push is O(1) operation
// In case of TTL and full channel O(n) worst case, where n is len of the channel
func (v *Vec) Push(w worker.BaseProcess) {
	// Non-blocking channel send
	select {
	case v.workers <- w:
		// default select branch is only possible when dealing with TTL
		// because in that case, workers in the v.workers channel can be TTL-ed and killed
		// but presenting in the channel
	default:
		v.Lock()
		defer v.Unlock()

		/*
			we can be in the default branch by the following reasons:
			1. TTL is set with no requests during the TTL
			2. Violated Get <-> Release operation (how ??)
		*/
		for i := uint64(0); i < v.len; i++ {
			wrk := <-v.workers
			switch wrk.State().Value() {
			// skip good states
			case worker.StateWorking, worker.StateReady:
				// put the worker back
				// generally, while send and receive operations are concurrent (from the channel), channel behave
				// like a FIFO, but when re-sending from the same goroutine it behaves like a FILO
				v.workers <- wrk
				continue
			default:
				// kill the current worker (just to be sure it's dead)
				_ = wrk.Kill()
				// replace with the new one
				v.workers <- w
				return
			}
		}
	}
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

	// used only for the TTL-ed workers
	v.RLock()
	defer v.RUnlock()

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
