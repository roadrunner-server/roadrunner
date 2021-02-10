package container

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/spiral/roadrunner/v2/pkg/worker"
)

type Stack struct {
	sync.RWMutex
	workers             []worker.BaseProcess
	destroy             bool
	actualNumOfWorkers  uint64
	initialNumOfWorkers uint64
}

func NewWorkersStack(initialNumOfWorkers uint64) *Stack {
	w := runtime.NumCPU()
	return &Stack{
		workers:             make([]worker.BaseProcess, 0, w),
		actualNumOfWorkers:  0,
		initialNumOfWorkers: initialNumOfWorkers,
	}
}

func (stack *Stack) Reset() {
	stack.Lock()
	defer stack.Unlock()
	stack.actualNumOfWorkers = 0
	stack.workers = nil
}

// Push worker back to the vec
// If vec in destroy state, Push will provide 100ms window to unlock the mutex
func (stack *Stack) Push(w worker.BaseProcess) {
	stack.Lock()
	defer stack.Unlock()
	stack.actualNumOfWorkers++
	stack.workers = append(stack.workers, w)
}

func (stack *Stack) IsEmpty() bool {
	stack.Lock()
	defer stack.Unlock()
	return len(stack.workers) == 0
}

func (stack *Stack) Pop() (worker.BaseProcess, bool) {
	stack.Lock()
	defer stack.Unlock()

	// do not release new vec
	if stack.destroy {
		return nil, true
	}

	if len(stack.workers) == 0 {
		return nil, false
	}

	// move worker
	w := stack.workers[len(stack.workers)-1]
	stack.workers = stack.workers[:len(stack.workers)-1]
	stack.actualNumOfWorkers--
	return w, false
}

func (stack *Stack) FindAndRemoveByPid(pid int64) bool {
	stack.Lock()
	defer stack.Unlock()
	for i := 0; i < len(stack.workers); i++ {
		// worker in the vec, reallocating
		if stack.workers[i].Pid() == pid {
			stack.workers = append(stack.workers[:i], stack.workers[i+1:]...)
			stack.actualNumOfWorkers--
			// worker found and removed
			return true
		}
	}
	// no worker with such ID
	return false
}

// Workers return copy of the workers in the vec
func (stack *Stack) Workers() []worker.BaseProcess {
	stack.Lock()
	defer stack.Unlock()
	workersCopy := make([]worker.BaseProcess, 0, 1)
	// copy
	// TODO pointers, copy have no sense
	for _, v := range stack.workers {
		if v != nil {
			workersCopy = append(workersCopy, v)
		}
	}

	return workersCopy
}

func (stack *Stack) isDestroying() bool {
	stack.Lock()
	defer stack.Unlock()
	return stack.destroy
}

// we also have to give a chance to pool to Push worker (return it)
func (stack *Stack) Destroy(_ context.Context) {
	stack.Lock()
	stack.destroy = true
	stack.Unlock()

	tt := time.NewTicker(time.Millisecond * 500)
	defer tt.Stop()
	for {
		select {
		case <-tt.C:
			stack.Lock()
			// that might be one of the workers is working
			if stack.initialNumOfWorkers != stack.actualNumOfWorkers {
				stack.Unlock()
				continue
			}
			stack.Unlock()
			// unnecessary mutex, but
			// just to make sure. All vec at this moment are in the vec
			// Pop operation is blocked, push can't be done, since it's not possible to pop
			stack.Lock()
			for i := 0; i < len(stack.workers); i++ {
				// set state for the vec in the vec (unused at the moment)
				stack.workers[i].State().Set(worker.StateDestroyed)
				// kill the worker
				_ = stack.workers[i].Kill()
			}
			stack.Unlock()
			// clear
			stack.Reset()
			return
		}
	}
}
