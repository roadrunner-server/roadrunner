package priorityqueue

type Queue interface {
	Push(item PQItem)
	Pop() PQItem
}

type PQItem interface {
	ID() string
	Priority() uint64
}
