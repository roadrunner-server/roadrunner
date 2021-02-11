package container

import (
	"sync/atomic"

	"github.com/spiral/roadrunner/v2/pkg/worker"
)

type Vec struct {
	destroy uint64
	workers chan worker.BaseProcess
}

func NewVector(initialNumOfWorkers uint64) Vector {
	vec := &Vec{
		destroy: 0,
		workers: make(chan worker.BaseProcess, initialNumOfWorkers),
	}

	return vec
}

func (v *Vec) Enqueue(w worker.BaseProcess) {
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

	w := <-v.workers

	return w, false
}

func (v *Vec) Destroy() {
	atomic.StoreUint64(&v.destroy, 1)
}
