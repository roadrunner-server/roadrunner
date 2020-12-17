package util

import (
	"sync"

	"github.com/spiral/roadrunner/v2/interfaces/events"
)

// EventHandler helps to broadcast events to multiple listeners.
type EventHandler struct {
	listeners []events.EventListener
	sync.Mutex
}

func NewEventsHandler() events.Handler {
	return &EventHandler{listeners: make([]events.EventListener, 0, 2)}
}

// NumListeners returns number of event listeners.
func (eb *EventHandler) NumListeners() int {
	eb.Lock()
	defer eb.Unlock()
	return len(eb.listeners)
}

// AddListener registers new event listener.
func (eb *EventHandler) AddListener(listener events.EventListener) {
	eb.Lock()
	defer eb.Unlock()
	eb.listeners = append(eb.listeners, listener)
}

// Push broadcast events across all event listeners.
func (eb *EventHandler) Push(e interface{}) {
	eb.Lock()
	defer eb.Unlock()
	for _, listener := range eb.listeners {
		listener(e)
	}
}
