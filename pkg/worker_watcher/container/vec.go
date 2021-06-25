package container

import (
	"sync/atomic"

	"github.com/spiral/roadrunner/v2/pkg/worker"
)

type Vec struct {
	wqLen   uint64
	destroy uint64
	workers chan worker.BaseProcess
}

func NewVector(initialNumOfWorkers uint64) Vector {
	vec := &Vec{
		wqLen:   0,
		destroy: 0,
		workers: make(chan worker.BaseProcess, initialNumOfWorkers),
	}

	return vec
}

func (v *Vec) Enqueue(w worker.BaseProcess) {
	atomic.AddUint64(&v.wqLen, 1)
	v.workers <- w
}

func (v *Vec) Dequeue() (worker.BaseProcess, bool) {
	/*
		if *addr == old {
			*addr = new
			return true
		}
	*/

	if atomic.CompareAndSwapUint64(&v.destroy, 1, 1) {
		return nil, true
	}

	if num := atomic.LoadUint64(&v.wqLen); num > 0 {
		atomic.AddUint64(&v.wqLen, ^uint64(0))
		return <-v.workers, false
	}

	return nil, false
}

func (v *Vec) Destroy() {
	atomic.StoreUint64(&v.destroy, 1)
}
