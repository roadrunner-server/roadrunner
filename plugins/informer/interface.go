package informer

import (
	"github.com/spiral/roadrunner/v2/pkg/process"
)

/*
Informer plugin should not receive any other plugin in the Init or via Collects
Because Availabler implementation should present in every plugin
*/

// Informer used to get workers from particular plugin or set of plugins
type Informer interface {
	Workers() []process.State
}

// Availabler interface should be implemented by every plugin which wish to report to the PHP worker that it available in the RR runtime
type Availabler interface {
	// Available method needed to collect all plugins which are available in the runtime.
	Available()
}
