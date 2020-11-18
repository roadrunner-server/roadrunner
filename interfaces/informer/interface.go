package informer

import "github.com/spiral/roadrunner/v2"

type Informer interface {
	Workers() []roadrunner.WorkerBase
}
