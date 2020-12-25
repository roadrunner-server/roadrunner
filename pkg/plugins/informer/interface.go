package informer

import "github.com/spiral/roadrunner/v2/interfaces/worker"

// Informer used to get workers from particular plugin or set of plugins
type Informer interface {
	Workers() []worker.BaseProcess
}
