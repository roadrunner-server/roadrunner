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
	workers []WorkerBase
	mutex   sync.RWMutex
	destroy bool
}

func NewWorkersStack() *Stack {
	return &Stack{
		workers: make([]WorkerBase, 0, runtime.NumCPU()),
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
	WorkersList() []WorkerBase

	// RemoveWorker remove worker from the stack
	RemoveWorker(ctx context.Context, wb WorkerBase) error
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

func (ww *workerWatcher) AddToWatch(ctx context.Context, workers []WorkerBase) error {
	for i := 0; i < len(workers); i++ {
		sw, err := NewSyncWorker(workers[i])
		if err != nil {
			return err
		}
		ww.stack.Push(sw)
		sw.AddListener(ww.events.Push)

		go func(swc WorkerBase) {
			ww.wait(ctx, swc)
		}(sw)
	}
	return nil
}

func (ww *workerWatcher) GetFreeWorker(ctx context.Context) (WorkerBase, error) {
	const op = errors.Op("get_free_worker")
	// thread safe operation
	w, stop := ww.stack.Pop()
	if stop {
		return nil, errors.E(op, errors.ErrWatcherStopped)
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
		// TODO allocate timeout
		tout := time.NewTicker(time.Second * 180)
		defer tout.Stop()
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
				ww.ReduceWorkersCount()
				return w, nil
			case <-tout.C:
				return nil, errors.E(op, errors.Str("no free workers in the stack"))
			}
		}
	}

	ww.ReduceWorkersCount()
	return w, nil
}

func (ww *workerWatcher) AllocateNew(ctx context.Context) error {
	ww.stack.mutex.Lock()
	const op = errors.Op("allocate new worker")
	sw, err := ww.allocator()
	if err != nil {
		return errors.E(op, err)
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

func (ww *workerWatcher) RemoveWorker(ctx context.Context, wb WorkerBase) error {
	ww.stack.mutex.Lock()
	const op = errors.Op("remove worker")
	defer ww.stack.mutex.Unlock()
	pid := wb.Pid()
	for i := 0; i < len(ww.stack.workers); i++ {
		if ww.stack.workers[i].Pid() == pid {
			// found in the stack
			// remove worker
			ww.stack.workers = append(ww.stack.workers[:i], ww.stack.workers[i+1:]...)
			ww.ReduceWorkersCount()

			wb.State().Set(StateInvalid)
			err := wb.Kill()
			if err != nil {
				return errors.E(op, err)
			}
			break
		}
	}
	// worker currently handle request, set state Remove
	wb.State().Set(StateRemove)
	return nil
}

// O(1) operation
func (ww *workerWatcher) PushWorker(w WorkerBase) {
	ww.IncreaseWorkersCount()
	ww.stack.Push(w)
}

func (ww *workerWatcher) ReduceWorkersCount() {
	ww.mutex.Lock()
	ww.actualNumWorkers--
	ww.mutex.Unlock()
}
func (ww *workerWatcher) IncreaseWorkersCount() {
	ww.mutex.Lock()
	ww.actualNumWorkers++
	ww.mutex.Unlock()
}

// Destroy all underlying stack (but let them to complete the task)
func (ww *workerWatcher) Destroy(ctx context.Context) {
	ww.stack.mutex.Lock()
	ww.stack.destroy = true
	ww.stack.mutex.Unlock()

	tt := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-tt.C:
			ww.stack.mutex.Lock()
			if len(ww.stack.workers) != int(ww.actualNumWorkers) {
				ww.stack.mutex.Unlock()
				continue
			}
			ww.stack.mutex.Unlock()
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

// Warning, this is O(n) operation, and it will return copy of the actual workers
func (ww *workerWatcher) WorkersList() []WorkerBase {
	ww.stack.mutex.Lock()
	defer ww.stack.mutex.Unlock()
	workersCopy := make([]WorkerBase, 0, 1)
	for _, v := range ww.stack.workers {
		sw := v.(SyncWorker)
		workersCopy = append(workersCopy, sw)
	}

	return workersCopy
}

func (ww *workerWatcher) wait(ctx context.Context, w WorkerBase) {
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

	pid := w.Pid()
	ww.stack.mutex.Lock()
	for i := 0; i < len(ww.stack.workers); i++ {
		// worker in the stack, reallocating
		if ww.stack.workers[i].Pid() == pid {
			ww.stack.workers = append(ww.stack.workers[:i], ww.stack.workers[i+1:]...)
			ww.ReduceWorkersCount()
			ww.stack.mutex.Unlock()

			err = ww.AllocateNew(ctx)
			if err != nil {
				ww.events.Push(PoolEvent{
					Event:   EventPoolError,
					Payload: errors.E(op, err),
				})
			}

			return
		}
	}

	ww.stack.mutex.Unlock()

	// worker not in the stack (not returned), forget and allocate new
	err = ww.AllocateNew(ctx)
	if err != nil {
		ww.events.Push(PoolEvent{
			Event:   EventPoolError,
			Payload: errors.E(op, err),
		})
		return
	}
}

func (ww *workerWatcher) addToWatch(wb WorkerBase) {
	ww.mutex.Lock()
	defer ww.mutex.Unlock()
	go func() {
		ww.wait(context.Background(), wb)
	}()
}
