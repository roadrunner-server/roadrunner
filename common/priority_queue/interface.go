package priorityqueue

type Queue interface {
	Insert(item Item)
	GetMax() Item
}

type Item interface {
	ID() string
	Priority() uint64
	Ack()
	Nack()
	Body() []byte
	Context() []byte
}
