package priorityqueue

type Queue interface {
	Push()
	Pop() interface{}
	BLPop()
}
