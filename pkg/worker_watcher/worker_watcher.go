package worker_watcher //nolint:golint,stylecheck

import (
	"context"
	"sync"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
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

func (ww *workerWatcher) Watch(workers []worker.SyncWorker) error {
	for i := 0; i < len(workers); i++ {
		ww.stack.Push(workers[i])

		go func(swc worker.SyncWorker) {
			ww.wait(swc)
		}(workers[i])
	}
	return nil
}

// Get is not a thread safe operation
func (ww *workerWatcher) Get(ctx context.Context) (worker.SyncWorker, error) {
	const op = errors.Op("worker_watcher_get_free_worker")
	// FAST PATH
	// thread safe operation
	w, stop := ww.stack.Pop()
	if stop {
		return nil, errors.E(op, errors.WatcherStopped)
	}

	// fast path, worker not nil and in the ReadyState
	if w != nil && w.State().Value() == worker.StateReady {
		return w, nil
	}
	// =========================================================
	// SLOW PATH
	// no free workers in the stack
	// try to continuously get free one
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

			switch w.State().Value() {
			case worker.StateRemove:
				err := ww.Remove(w)
				if err != nil {
					return nil, errors.E(op, err)
				}
				// try to get next
				continue
			case
				// all the possible wrong states
				worker.StateInactive,
				worker.StateDestroyed,
				worker.StateErrored,
				worker.StateStopped,
				worker.StateInvalid,
				worker.StateKilling,
				worker.StateWorking, // ??? how
				worker.StateStopping:
				// worker doing no work because it in the stack
				// so we can safely kill it (inconsistent state)
				_ = w.Kill()
				// try to get new worker
				continue
				// return only workers in the Ready state
			case worker.StateReady:
				return w, nil
			}
		case <-ctx.Done():
			return nil, errors.E(op, errors.NoFreeWorkers, errors.Str("no free workers in the stack, timeout exceed"))
		}
	}
}

func (ww *workerWatcher) Allocate() error {
	ww.stack.mutex.Lock()
	const op = errors.Op("worker_watcher_allocate_new")
	sw, err := ww.allocator()
	if err != nil {
		return errors.E(op, errors.WorkerAllocate, err)
	}

	ww.addToWatch(sw)
	ww.stack.mutex.Unlock()
	ww.Push(sw)

	return nil
}

// Remove
func (ww *workerWatcher) Remove(wb worker.SyncWorker) error {
	ww.mutex.Lock()
	defer ww.mutex.Unlock()

	const op = errors.Op("worker_watcher_remove_worker")
	// set remove state
	wb.State().Set(worker.StateRemove)
	if ww.stack.FindAndRemoveByPid(wb.Pid()) {
		err := wb.Kill()
		if err != nil {
			return errors.E(op, err)
		}
		return nil
	}

	return nil
}

// O(1) operation
func (ww *workerWatcher) Push(w worker.SyncWorker) {
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
func (ww *workerWatcher) List() []worker.SyncWorker {
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

	if w.State().Value() == worker.StateDestroyed {
		// worker was manually destroyed, no need to replace
		ww.events.Push(events.PoolEvent{Event: events.EventWorkerDestruct, Payload: w})
		return
	}

	_ = ww.stack.FindAndRemoveByPid(w.Pid())
	err = ww.Allocate()
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
