package worker_watcher //nolint:golint,stylecheck

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/interfaces/events"
	"github.com/spiral/roadrunner/v2/interfaces/worker"
	"github.com/spiral/roadrunner/v2/internal"
)

type Stack struct {
	workers            []worker.BaseProcess
	mutex              sync.RWMutex
	destroy            bool
	actualNumOfWorkers int64
}

func NewWorkersStack() *Stack {
	w := runtime.NumCPU()
	return &Stack{
		workers:            make([]worker.BaseProcess, 0, w),
		actualNumOfWorkers: 0,
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
	stack.workers = append(stack.workers, w)
}

func (stack *Stack) IsEmpty() bool {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	return len(stack.workers) == 0
}

func (stack *Stack) Pop() (worker.BaseProcess, bool) {
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
func (stack *Stack) Workers() []worker.BaseProcess {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	workersCopy := make([]worker.BaseProcess, 0, 1)
	// copy
	for _, v := range stack.workers {
		workersCopy = append(workersCopy, v)
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

	tt := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-tt.C:
			stack.mutex.Lock()
			// that might be one of the workers is working
			if len(stack.workers) != int(stack.actualNumOfWorkers) {
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
				stack.workers[i].State().Set(internal.StateDestroyed)
				// kill the worker
				_ = stack.workers[i].Kill()
			}
			stack.mutex.Unlock()
			tt.Stop()
			// clear
			stack.Reset()
			return
		}
	}
}

// workerCreateFunc can be nil, but in that case, dead stack will not be replaced
func NewWorkerWatcher(allocator worker.Allocator, numWorkers int64, events events.Handler) worker.Watcher {
	ww := &workerWatcher{
		stack:             NewWorkersStack(),
		allocator:         allocator,
		initialNumWorkers: numWorkers,
		actualNumWorkers:  numWorkers,
		events:            events,
	}

	return ww
}

type workerWatcher struct {
	mutex             sync.RWMutex
	stack             *Stack
	allocator         worker.Allocator
	initialNumWorkers int64
	actualNumWorkers  int64
	events            events.Handler
}

func (ww *workerWatcher) AddToWatch(workers []worker.BaseProcess) error {
	for i := 0; i < len(workers); i++ {
		ww.stack.Push(workers[i])

		go func(swc worker.BaseProcess) {
			ww.wait(swc)
		}(workers[i])
	}
	return nil
}

func (ww *workerWatcher) GetFreeWorker(ctx context.Context) (worker.BaseProcess, error) {
	const op = errors.Op("GetFreeWorker")
	// thread safe operation
	w, stop := ww.stack.Pop()
	if stop {
		return nil, errors.E(op, errors.ErrWatcherStopped)
	}

	// handle worker remove state
	// in this state worker is destroyed by supervisor
	if w != nil && w.State().Value() == internal.StateRemove {
		err := ww.RemoveWorker(w)
		if err != nil {
			return nil, err
		}
		// try to get next
		return ww.GetFreeWorker(ctx)
	}
	// no free stack
	if w == nil {
		for {
			select {
			default:
				w, stop = ww.stack.Pop()
				if stop {
					return nil, errors.E(op, errors.ErrWatcherStopped)
				}
				if w == nil {
					continue
				}
				return w, nil
			case <-ctx.Done():
				return nil, errors.E(op, errors.NoFreeWorkers, errors.Str("no free workers in the stack, timeout exceed"))
			}
		}
	}

	return w, nil
}

func (ww *workerWatcher) AllocateNew() error {
	ww.stack.mutex.Lock()
	const op = errors.Op("allocate new worker")
	sw, err := ww.allocator()
	if err != nil {
		return errors.E(op, errors.WorkerAllocate, err)
	}

	ww.addToWatch(sw)
	ww.stack.mutex.Unlock()
	ww.PushWorker(sw)

	return nil
}

func (ww *workerWatcher) RemoveWorker(wb worker.BaseProcess) error {
	ww.mutex.Lock()
	defer ww.mutex.Unlock()

	const op = errors.Op("remove worker")
	pid := wb.Pid()

	if ww.stack.FindAndRemoveByPid(pid) {
		wb.State().Set(internal.StateInvalid)
		err := wb.Kill()
		if err != nil {
			return errors.E(op, err)
		}
		return nil
	}

	wb.State().Set(internal.StateRemove)
	return nil
}

// O(1) operation
func (ww *workerWatcher) PushWorker(w worker.BaseProcess) {
	ww.mutex.Lock()
	defer ww.mutex.Unlock()
	ww.stack.Push(w)
}

// Destroy all underlying stack (but let them to complete the task)
func (ww *workerWatcher) Destroy(ctx context.Context) {
	// destroy stack, we don't use ww mutex here, since we should be able to push worker
	ww.stack.Destroy(ctx)
}

// Warning, this is O(n) operation, and it will return copy of the actual workers
func (ww *workerWatcher) WorkersList() []worker.BaseProcess {
	return ww.stack.Workers()
}

func (ww *workerWatcher) wait(w worker.BaseProcess) {
	const op = errors.Op("process wait")
	err := w.Wait()
	if err != nil {
		ww.events.Push(events.WorkerEvent{
			Event:   events.EventWorkerError,
			Worker:  w,
			Payload: errors.E(op, err),
		})
	}

	if w.State().Value() == internal.StateDestroyed {
		// worker was manually destroyed, no need to replace
		ww.events.Push(events.PoolEvent{Event: events.EventWorkerDestruct, Payload: w})
		return
	}

	_ = ww.stack.FindAndRemoveByPid(w.Pid())
	err = ww.AllocateNew()
	if err != nil {
		ww.events.Push(events.PoolEvent{
			Event:   events.EventPoolError,
			Payload: errors.E(op, err),
		})
	}
}

func (ww *workerWatcher) addToWatch(wb worker.BaseProcess) {
	go func() {
		ww.wait(wb)
	}()
}
