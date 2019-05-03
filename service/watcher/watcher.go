package watcher

import "github.com/spiral/roadrunner"

// Watchable defines the ability to attach roadrunner watcher.
type Watchable interface {
	// Watch attaches watcher to the service.
	Watch(w roadrunner.Watcher)
}
