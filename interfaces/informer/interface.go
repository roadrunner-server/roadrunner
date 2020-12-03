package informer

import "github.com/spiral/roadrunner/v2"

// Informer used to get workers from particular plugin or set of plugins
type Informer interface {
	Workers() []roadrunner.WorkerBase
}
