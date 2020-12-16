package util

import "github.com/spiral/roadrunner/v2/interfaces/worker"

// EventHandler helps to broadcast events to multiple listeners.
type EventHandler struct {
	listeners []worker.EventListener
}

func NewEventsHandler() worker.EventsHandler {
	return &EventHandler{listeners: make([]worker.EventListener, 0, 2)}
}

// NumListeners returns number of event listeners.
func (eb *EventHandler) NumListeners() int {
	return len(eb.listeners)
}

// AddListener registers new event listener.
func (eb *EventHandler) AddListener(listener worker.EventListener) {
	eb.listeners = append(eb.listeners, listener)
}

// Push broadcast events across all event listeners.
func (eb *EventHandler) Push(e interface{}) {
	for _, listener := range eb.listeners {
		listener(e)
	}
}
