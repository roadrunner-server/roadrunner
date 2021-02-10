package container

import "github.com/spiral/roadrunner/v2/pkg/worker"

type Vector interface {
	Enqueue(worker.BaseProcess)
	Dequeue() (worker.BaseProcess, bool)
	Destroy()
}
