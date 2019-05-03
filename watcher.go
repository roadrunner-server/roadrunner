package roadrunner

// Watcher observes pool state and decides if any worker must be destroyed.
type Watcher interface {
	// Lock watcher on given pool instance.
	Attach(p Pool) Watcher

	// Detach pool watching.
	Detach()
}
