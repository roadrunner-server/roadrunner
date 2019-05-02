package roadrunner

import (
	"sync"
	"time"
)

// Watcher watches for workers.
type Watcher interface {
	// Keep must return true and nil if worker is OK to continue working,
	// must return false and optional error to force worker destruction.
	Keep(p Pool, w *Worker) (keep bool, err error)
}

// disconnect??
type LazyWatcher struct {
	// defines how often
	interval time.Duration

	mu sync.Mutex
	p  Pool
}
