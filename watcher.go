package roadrunner

// Watcher watches for workers.
type Watcher interface {
	// Keep must return true and nil if worker is OK to continue working,
	// must return false and optional error to force worker destruction.
	Keep(p Pool, w *Worker) (keep bool, err error)
}
