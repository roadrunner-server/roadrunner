package worker_watcher //nolint:golint,stylecheck

import (
	"context"
	"sync"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/states"
	"github.com/spiral/roadrunner/v2/pkg/worker"
)

// workerCreateFunc can be nil, but in that case, dead stack will not be replaced
func NewSyncWorkerWatcher(allocator worker.Allocator, numWorkers uint64, events events.Handler) Watcher {
	ww := &workerWatcher{
		stack:     NewWorkersStack(numWorkers),
		allocator: allocator,
		events:    events,
	}

	return ww
}

type workerWatcher struct {
	mutex     sync.RWMutex
	stack     *Stack
	allocator worker.Allocator
	events    events.Handler
}

func (ww *workerWatcher) AddToWatch(workers []worker.SyncWorker) error {
	for i := 0; i < len(workers); i++ {
		ww.stack.Push(workers[i])

		go func(swc worker.SyncWorker) {
			ww.wait(swc)
		}(workers[i])
	}
	return nil
}

func (ww *workerWatcher) GetFreeWorker(ctx context.Context) (worker.SyncWorker, error) {
	const op = errors.Op("worker_watcher_get_free_worker")
	// thread safe operation
	w, stop := ww.stack.Pop()
	if stop {
		return nil, errors.E(op, errors.WatcherStopped)
	}

	// handle worker remove state
	// in this state worker is destroyed by supervisor
	if w != nil && w.State().Value() == states.StateRemove {
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
					return nil, errors.E(op, errors.WatcherStopped)
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
	const op = errors.Op("worker_watcher_allocate_new")
	sw, err := ww.allocator()
	if err != nil {
		return errors.E(op, errors.WorkerAllocate, err)
	}

	ww.addToWatch(sw)
	ww.stack.mutex.Unlock()
	ww.PushWorker(sw)

	return nil
}

func (ww *workerWatcher) RemoveWorker(wb worker.SyncWorker) error {
	ww.mutex.Lock()
	defer ww.mutex.Unlock()

	const op = errors.Op("worker_watcher_remove_worker")
	pid := wb.Pid()

	if ww.stack.FindAndRemoveByPid(pid) {
		wb.State().Set(states.StateRemove)
		err := wb.Kill()
		if err != nil {
			return errors.E(op, err)
		}
		return nil
	}

	return nil
}

// O(1) operation
func (ww *workerWatcher) PushWorker(w worker.SyncWorker) {
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
func (ww *workerWatcher) WorkersList() []worker.SyncWorker {
	return ww.stack.Workers()
}

func (ww *workerWatcher) wait(w worker.BaseProcess) {
	const op = errors.Op("worker_watcher_wait")
	err := w.Wait()
	if err != nil {
		ww.events.Push(events.WorkerEvent{
			Event:   events.EventWorkerError,
			Worker:  w,
			Payload: errors.E(op, err),
		})
	}

	if w.State().Value() == states.StateDestroyed {
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

func (ww *workerWatcher) addToWatch(wb worker.SyncWorker) {
	go func() {
		ww.wait(wb)
	}()
}
