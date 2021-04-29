package informer

import (
	"github.com/spiral/roadrunner/v2/pkg/process"
)

// Informer used to get workers from particular plugin or set of plugins
type Informer interface {
	Workers() []process.State
}

// Lister interface used to filter available plugins
type Lister interface {
	// List gets no args, but returns list of the active plugins
	List() []string
}
