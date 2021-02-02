package worker_watcher //nolint:golint,stylecheck
import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/spiral/roadrunner/v2/pkg/states"
	"github.com/spiral/roadrunner/v2/pkg/worker"
)

type Stack struct {
	workers             []*worker.SyncWorkerImpl
	mutex               sync.RWMutex
	destroy             bool
	actualNumOfWorkers  uint64
	initialNumOfWorkers uint64
}

func NewWorkersStack(initialNumOfWorkers uint64) *Stack {
	w := runtime.NumCPU()
	return &Stack{
		workers:             make([]*worker.SyncWorkerImpl, 0, w),
		actualNumOfWorkers:  0,
		initialNumOfWorkers: initialNumOfWorkers,
	}
}

func (stack *Stack) Reset() {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	stack.actualNumOfWorkers = 0
	stack.workers = nil
}

// Push worker back to the stack
// If stack in destroy state, Push will provide 100ms window to unlock the mutex
func (stack *Stack) Push(w worker.BaseProcess) {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	stack.actualNumOfWorkers++
	stack.workers = append(stack.workers, w.(*worker.SyncWorkerImpl))
}

func (stack *Stack) IsEmpty() bool {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	return len(stack.workers) == 0
}

func (stack *Stack) Pop() (*worker.SyncWorkerImpl, bool) {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()

	// do not release new stack
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
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	for i := 0; i < len(stack.workers); i++ {
		// worker in the stack, reallocating
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

// Workers return copy of the workers in the stack
func (stack *Stack) Workers() []worker.SyncWorker {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	workersCopy := make([]worker.SyncWorker, 0, 1)
	// copy
	for _, v := range stack.workers {
		if v != nil {
			workersCopy = append(workersCopy, v)
		}
	}

	return workersCopy
}

func (stack *Stack) isDestroying() bool {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	return stack.destroy
}

// we also have to give a chance to pool to Push worker (return it)
func (stack *Stack) Destroy(ctx context.Context) {
	stack.mutex.Lock()
	stack.destroy = true
	stack.mutex.Unlock()

	tt := time.NewTicker(time.Millisecond * 500)
	defer tt.Stop()
	for {
		select {
		case <-tt.C:
			stack.mutex.Lock()
			// that might be one of the workers is working
			if stack.initialNumOfWorkers != stack.actualNumOfWorkers {
				stack.mutex.Unlock()
				continue
			}
			stack.mutex.Unlock()
			// unnecessary mutex, but
			// just to make sure. All stack at this moment are in the stack
			// Pop operation is blocked, push can't be done, since it's not possible to pop
			stack.mutex.Lock()
			for i := 0; i < len(stack.workers); i++ {
				// set state for the stack in the stack (unused at the moment)
				stack.workers[i].State().Set(states.StateDestroyed)
				// kill the worker
				_ = stack.workers[i].Kill()
			}
			stack.mutex.Unlock()
			// clear
			stack.Reset()
			return
		}
	}
}
