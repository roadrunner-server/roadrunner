package events

// Handler interface
type Handler interface {
	// Return number of active listeners
	NumListeners() int
	// AddListener adds lister to the publisher
	AddListener(listener Listener)
	// Push pushes event to the listeners
	Push(e interface{})
}

// Event listener listens for the events produced by worker, worker pool or other service.
type Listener func(event interface{})
