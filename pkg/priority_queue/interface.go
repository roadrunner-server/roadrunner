package priorityqueue

type Queue interface {
	Push()
	Pop()
	BLPop()
}
