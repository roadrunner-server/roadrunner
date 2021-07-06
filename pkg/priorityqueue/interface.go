package priorityqueue

type Queue interface {
	Insert(item Item)
	GetMax() Item
}

type Item interface {
	ID() *string
	Priority() *uint64
	Body() []byte
	Context() []byte
	Ack()
	Nack()
}
