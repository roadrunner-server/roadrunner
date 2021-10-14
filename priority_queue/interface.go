package priorityqueue

type Queue interface {
	Insert(item Item)
	ExtractMin() Item
	Len() uint64
}

// Item represents binary heap item
type Item interface {
	// ID is a unique item identifier
	ID() string

	// Priority returns the Item's priority to sort
	Priority() int64

	// Body is the Item payload
	Body() []byte

	// Context is the Item meta information
	Context() ([]byte, error)
}
