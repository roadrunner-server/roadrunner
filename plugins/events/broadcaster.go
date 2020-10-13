package events

type EventListener interface {
	Handle(event interface{})
}

type EventBroadcaster struct {
	listeners []EventListener
}

func NewEventBroadcaster() *EventBroadcaster {
	return &EventBroadcaster{}
}

func (eb *EventBroadcaster) AddListener(l EventListener) {
	// todo: threadcase
	eb.listeners = append(eb.listeners, l)
}

func (eb *EventBroadcaster) Push(e interface{}) {
	for _, l := range eb.listeners {
		l.Handle(e)
	}
}
