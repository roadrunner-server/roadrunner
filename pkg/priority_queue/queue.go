package priorityqueue

import "fmt"

type QueueImpl struct {
}

func NewPriorityQueue() *QueueImpl {
	return &QueueImpl{}
}

// Push the task
func (q *QueueImpl) Push(item interface{}) {
	fmt.Println(item)
}

func (q *QueueImpl) Pop() interface{} {
	return nil
}
