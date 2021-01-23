package events

import (
	"sync"
)

// HandlerImpl helps to broadcast events to multiple listeners.
type HandlerImpl struct {
	listeners    []Listener
	sync.RWMutex // all receivers should be pointers
}

func NewEventsHandler() Handler {
	return &HandlerImpl{listeners: make([]Listener, 0, 2)}
}

// NumListeners returns number of event listeners.
func (eb *HandlerImpl) NumListeners() int {
	eb.Lock()
	defer eb.Unlock()
	return len(eb.listeners)
}

// AddListener registers new event listener.
func (eb *HandlerImpl) AddListener(listener Listener) {
	eb.Lock()
	defer eb.Unlock()
	eb.listeners = append(eb.listeners, listener)
}

// Push broadcast events across all event listeners.
func (eb *HandlerImpl) Push(e interface{}) {
	// ReadLock here because we are not changing listeners
	eb.RLock()
	defer eb.RUnlock()
	for k := range eb.listeners {
		eb.listeners[k](e)
	}
}
