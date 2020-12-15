package roadrunner

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/util"
)

type Stack struct {
	workers            []WorkerBase
	mutex              sync.RWMutex
	destroy            bool
	actualNumOfWorkers int64
}

func NewWorkersStack() *Stack {
	w := runtime.NumCPU()
	return &Stack{
		workers:            make([]WorkerBase, 0, w),
		actualNumOfWorkers: 0,
	}
}

func (stack *Stack) Reset() {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	stack.actualNumOfWorkers = 0
	stack.workers = nil
}

func (stack *Stack) Push(w WorkerBase) {
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

func (stack *Stack) Pop() (WorkerBase, bool) {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	// do not release new stack
	if stack.destroy {
		return nil, true
	}

	if len(stack.workers) == 0 {
		return nil, false
	}

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

func (stack *Stack) Workers() []WorkerBase {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	workersCopy := make([]WorkerBase, 0, 1)
	// copy
	for _, v := range stack.workers {
		sw := v.(SyncWorker)
		workersCopy = append(workersCopy, sw)
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
				stack.workers[i].State().Set(StateDestroyed)
			}
			stack.mutex.Unlock()
			tt.Stop()
			// clear
			stack.Reset()
			return
		}
	}
}

type WorkerWatcher interface {
	// AddToWatch used to add stack to wait its state
	AddToWatch(workers []WorkerBase) error

	// GetFreeWorker provide first free worker
	GetFreeWorker(ctx context.Context) (WorkerBase, error)

	// PutWorker enqueues worker back
	PushWorker(w WorkerBase)

	// AllocateNew used to allocate new worker and put in into the WorkerWatcher
	AllocateNew() error

	// Destroy destroys the underlying stack
	Destroy(ctx context.Context)

	// WorkersList return all stack w/o removing it from internal storage
	WorkersList() []WorkerBase

	// RemoveWorker remove worker from the stack
	RemoveWorker(wb WorkerBase) error
}

// workerCreateFunc can be nil, but in that case, dead stack will not be replaced
func newWorkerWatcher(allocator Allocator, numWorkers int64, events util.EventsHandler) WorkerWatcher {
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
	allocator         Allocator
	initialNumWorkers int64
	actualNumWorkers  int64
	events            util.EventsHandler
}

func (ww *workerWatcher) AddToWatch(workers []WorkerBase) error {
	for i := 0; i < len(workers); i++ {
		sw, err := NewSyncWorker(workers[i])
		if err != nil {
			return err
		}
		ww.stack.Push(sw)
		sw.AddListener(ww.events.Push)

		go func(swc WorkerBase) {
			ww.wait(swc)
		}(sw)
	}
	return nil
}

func (ww *workerWatcher) GetFreeWorker(ctx context.Context) (WorkerBase, error) {
	const op = errors.Op("GetFreeWorker")
	// thread safe operation
	w, stop := ww.stack.Pop()
	if stop {
		return nil, errors.E(op, errors.ErrWatcherStopped)
	}

	// handle worker remove state
	// in this state worker is destroyed by supervisor
	if w != nil && w.State().Value() == StateRemove {
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
				//ww.ReduceWorkersCount()
				return w, nil
			case <-ctx.Done():
				return nil, errors.E(op, errors.NoFreeWorkers, errors.Str("no free workers in the stack, timeout exceed"))
			}
		}
	}

	//ww.ReduceWorkersCount()
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

	ww.events.Push(PoolEvent{
		Event:   EventWorkerConstruct,
		Payload: sw,
	})

	return nil
}

func (ww *workerWatcher) RemoveWorker(wb WorkerBase) error {
	ww.mutex.Lock()
	defer ww.mutex.Unlock()

	const op = errors.Op("remove worker")
	pid := wb.Pid()

	if ww.stack.FindAndRemoveByPid(pid) {
		wb.State().Set(StateInvalid)
		err := wb.Kill()
		if err != nil {
			return errors.E(op, err)
		}
		return nil
	}

	wb.State().Set(StateRemove)
	return nil

}

// O(1) operation
func (ww *workerWatcher) PushWorker(w WorkerBase) {
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
func (ww *workerWatcher) WorkersList() []WorkerBase {
	return ww.stack.Workers()
}

func (ww *workerWatcher) wait(w WorkerBase) {
	const op = errors.Op("process wait")
	err := w.Wait()
	if err != nil {
		ww.events.Push(WorkerEvent{
			Event:   EventWorkerError,
			Worker:  w,
			Payload: errors.E(op, err),
		})
	}

	if w.State().Value() == StateDestroyed {
		// worker was manually destroyed, no need to replace
		return
	}

	_ = ww.stack.FindAndRemoveByPid(w.Pid())
	err = ww.AllocateNew()
	if err != nil {
		ww.events.Push(PoolEvent{
			Event:   EventPoolError,
			Payload: errors.E(op, err),
		})
	}
}

func (ww *workerWatcher) addToWatch(wb WorkerBase) {
	ww.mutex.Lock()
	defer ww.mutex.Unlock()
	go func() {
		ww.wait(wb)
	}()
}
