package roadrunner

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrWatcherStopped = errors.New("watcher stopped")

type Stack struct {
	workers []WorkerBase
	mutex   sync.RWMutex
	destroy bool
}

func NewWorkersStack() *Stack {
	return &Stack{
		workers: make([]WorkerBase, 0, 12),
	}
}

func (stack *Stack) Reset() {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()

	stack.workers = nil
}

func (stack *Stack) Push(w WorkerBase) {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
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

	return w, false
}

type WorkersWatcher struct {
	mutex             sync.Mutex
	stack             *Stack
	allocator         func(args ...interface{}) (*SyncWorker, error)
	initialNumWorkers int64
	actualNumWorkers  int64
	events            chan PoolEvent
}

type WorkerWatcher interface {
	// AddToWatch used to add stack to wait its state
	AddToWatch(ctx context.Context, workers []WorkerBase) error
	// GetFreeWorker provide first free worker
	GetFreeWorker(ctx context.Context) (WorkerBase, error)
	// PutWorker enqueues worker back
	PushWorker(w WorkerBase)
	// AllocateNew used to allocate new worker and put in into the WorkerWatcher
	AllocateNew(ctx context.Context) error
	// Destroy destroys the underlying stack
	Destroy(ctx context.Context)
	// WorkersList return all stack w/o removing it from internal storage
	WorkersList(ctx context.Context) []WorkerBase
	// RemoveWorker remove worker from the stack
	RemoveWorker(ctx context.Context, wb WorkerBase) error
}

// workerCreateFunc can be nil, but in that case, dead stack will not be replaced
func NewWorkerWatcher(allocator func(args ...interface{}) (*SyncWorker, error), numWorkers int64, events chan PoolEvent) *WorkersWatcher {
	// todo check if events not nil
	ww := &WorkersWatcher{
		stack:             NewWorkersStack(),
		allocator:         allocator,
		initialNumWorkers: numWorkers,
		actualNumWorkers:  numWorkers,
		events:            events,
	}

	return ww
}

func (ww *WorkersWatcher) AddToWatch(ctx context.Context, workers []WorkerBase) error {
	for i := 0; i < len(workers); i++ {
		sw, err := NewSyncWorker(workers[i])
		if err != nil {
			return err
		}
		ww.stack.Push(sw)
		go func(swc WorkerBase) {
			//ww.mutex.Lock()
			ww.watch(&swc)
			ww.wait(ctx, &swc)
			//ww.mutex.Unlock()
		}(sw)
	}
	return nil
}

func (ww *WorkersWatcher) GetFreeWorker(ctx context.Context) (WorkerBase, error) {
	// thread safe operation
	w, stop := ww.stack.Pop()
	if stop {
		return nil, ErrWatcherStopped
	}
	// handle worker remove state
	// in this state worker is destroyed by supervisor
	if w != nil && w.State().Value() == StateRemove {
		err := ww.RemoveWorker(ctx, w)
		if err != nil {
			return nil, err
		}
		// try to get next
		return ww.GetFreeWorker(ctx)
	}
	// no free stack
	if w == nil {
		tout := time.NewTicker(time.Second * 180)
		defer tout.Stop()
		for {
			select {
			default:
				w, stop = ww.stack.Pop()
				if stop {
					return nil, ErrWatcherStopped
				}
				if w == nil {
					continue
				}
				ww.actualNumWorkers--
				return w, nil
			case <-tout.C:
				return nil, errors.New("no free stack")
			}
		}
	}
	ww.actualNumWorkers--
	return w, nil
}

func (ww *WorkersWatcher) AllocateNew(ctx context.Context) error {
	ww.stack.mutex.Lock()
	sw, err := ww.allocator()
	if err != nil {
		return err
	}
	ww.addToWatch(*sw)
	ww.stack.mutex.Unlock()
	ww.PushWorker(*sw)
	return nil
}

func (ww *WorkersWatcher) RemoveWorker(ctx context.Context, wb WorkerBase) error {
	ww.stack.mutex.Lock()
	defer ww.stack.mutex.Unlock()
	pid := wb.Pid()
	for i := 0; i < len(ww.stack.workers); i++ {
		if ww.stack.workers[i].Pid() == pid {
			// found in the stack
			// remove worker
			ww.stack.workers = append(ww.stack.workers[:i], ww.stack.workers[i+1:]...)
			ww.decreaseNumOfActualWorkers()

			wb.State().Set(StateInvalid)
			err := wb.Kill(ctx)
			if err != nil {
				return err
			}
			break
		}
	}
	// worker currently handle request, set state Remove
	wb.State().Set(StateRemove)
	return nil
}

// O(1) operation
func (ww *WorkersWatcher) PushWorker(w WorkerBase) {
	ww.mutex.Lock()
	ww.actualNumWorkers++
	ww.mutex.Unlock()
	ww.stack.Push(w)
}

func (ww *WorkersWatcher) ReduceWorkersCount() {
	ww.mutex.Unlock()
	ww.actualNumWorkers--
	ww.mutex.Lock()
}

// Destroy all underlying stack (but let them to complete the task)
func (ww *WorkersWatcher) Destroy(ctx context.Context) {
	ww.stack.mutex.Lock()
	ww.stack.destroy = true
	ww.stack.mutex.Unlock()

	tt := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-tt.C:
			if len(ww.stack.workers) != int(ww.actualNumWorkers) {
				continue
			}
			// unnecessary mutex, but
			// just to make sure. All stack at this moment are in the stack
			// Pop operation is blocked, push can't be done, since it's not possible to pop
			ww.stack.mutex.Lock()
			for i := 0; i < len(ww.stack.workers); i++ {
				// set state for the stack in the stack (unused at the moment)
				ww.stack.workers[i].State().Set(StateDestroyed)
			}
			ww.stack.mutex.Unlock()
			tt.Stop()
			// clear
			ww.stack.Reset()
			return
		}
	}
}

// Warning, this is O(n) operation
func (ww *WorkersWatcher) WorkersList(ctx context.Context) []WorkerBase {
	return ww.stack.workers
}

func (ww *WorkersWatcher) wait(ctx context.Context, w *WorkerBase) {
	err := (*w).Wait(ctx)
	if err != nil {
		ww.events <- PoolEvent{Payload: WorkerEvent{
			Event:   EventWorkerError,
			Worker:  *w,
			Payload: err,
		}}
	}
	// If not destroyed, reallocate
	if (*w).State().Value() != StateDestroyed {
		pid := (*w).Pid()
		for i := 0; i < len(ww.stack.workers); i++ {
			// worker in the stack, reallocating
			if ww.stack.workers[i].Pid() == pid {
				ww.stack.mutex.Lock()
				ww.stack.workers = append(ww.stack.workers[:i], ww.stack.workers[i+1:]...)

				ww.decreaseNumOfActualWorkers()

				ww.stack.mutex.Unlock()
				err = ww.AllocateNew(ctx)
				if err != nil {
					ww.events <- PoolEvent{Payload: WorkerEvent{
						Event:   EventWorkerError,
						Worker:  *w,
						Payload: err,
					}}
					return
				}
				return
			}
		}
		// worker not in the stack (not returned), forget and allocate new
		err = ww.AllocateNew(ctx)
		if err != nil {
			ww.events <- PoolEvent{Payload: WorkerEvent{
				Event:   EventWorkerError,
				Worker:  *w,
				Payload: err,
			}}
			return
		}
	}
	return
}

func (ww *WorkersWatcher) addToWatch(wb WorkerBase) {
	ww.mutex.Lock()
	defer ww.mutex.Unlock()
	go func() {
		ww.wait(context.Background(), &wb)
	}()
}

func (ww *WorkersWatcher) reallocate(wb *WorkerBase) error {
	sw, err := ww.allocator()
	if err != nil {
		return err
	}
	*wb = *sw
	return nil
}

func (ww *WorkersWatcher) decreaseNumOfActualWorkers() {
	ww.mutex.Lock()
	ww.actualNumWorkers--
	ww.mutex.Unlock()
}

func (ww *WorkersWatcher) watch(swc *WorkerBase) {
	// todo make event to stop function
	go func() {
		select {
		case ev := <-(*swc).Events():
			ww.events <- PoolEvent{Payload: ev}
		}
	}()
}
