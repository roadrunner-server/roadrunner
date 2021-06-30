package priorityqueue

type Queue interface {
	Push(item interface{})
	Pop() interface{}
}
