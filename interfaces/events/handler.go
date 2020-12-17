package events

type Handler interface {
	NumListeners() int
	AddListener(listener EventListener)
	Push(e interface{})
}

// Event listener listens for the events produced by worker, worker pool or other service.
type EventListener func(event interface{})
