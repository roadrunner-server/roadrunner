package priorityqueue

type Queue interface {
	Insert(item PQItem)
	GetMax() PQItem
}

type PQItem interface {
	ID() string
	Priority() uint64
}
