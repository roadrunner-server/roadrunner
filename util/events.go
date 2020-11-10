package util

type EventsHandler interface {
	NumListeners() int
	AddListener(listener EventListener)
	Push(e interface{})
}

// Event listener listens for the events produced by worker, worker pool or other servce.
type EventListener func(event interface{})

// EventHandler helps to broadcast events to multiple listeners.
type EventHandler struct {
	listeners []EventListener
}

func NewEventsHandler() EventsHandler {
	return &EventHandler{listeners: make([]EventListener, 0, 2)}
}

// NumListeners returns number of event listeners.
func (eb *EventHandler) NumListeners() int {
	return len(eb.listeners)
}

// AddListener registers new event listener.
func (eb *EventHandler) AddListener(listener EventListener) {
	eb.listeners = append(eb.listeners, listener)
}

// Push broadcast events across all event listeners.
func (eb *EventHandler) Push(e interface{}) {
	for _, listener := range eb.listeners {
		listener(e)
	}
}
