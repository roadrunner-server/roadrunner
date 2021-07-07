package priorityqueue

type Queue interface {
	Insert(item Item)
	GetMax() Item
}

// Item represents binary heap item
type Item interface {
	// ID is a unique item identifier
	ID() string

	// Priority returns the Item's priority to sort
	Priority() uint64

	// Body is the Item payload
	Body() []byte

	// Context is the Item meta information
	Context() ([]byte, error)

	// Ack - acknowledge the Item after processing
	Ack()

	// Nack - discard the Item
	Nack()
}
