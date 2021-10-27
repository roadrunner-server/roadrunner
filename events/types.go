package events

type EventBus interface {
	SubscribeAll(subID string, ch chan<- Event) error
	SubscribeP(subID string, pattern string, ch chan<- Event) error
	Unsubscribe(subID string)
	UnsubscribeP(subID, pattern string)
	Len() uint
	Send(ev Event)
}

type Event interface {
	Plugin() string
	Type() EventType
	Message() string
}

type RREvent struct {
	// event typ
	typ EventType
	// plugin
	plugin string
	// message
	message string
}

// NewEvent initializes new event
func NewEvent(t EventType, plugin string, msg string) *RREvent {
	return &RREvent{
		typ:     t,
		plugin:  plugin,
		message: msg,
	}
}

func (r *RREvent) Type() EventType {
	return r.typ
}

func (r *RREvent) Message() string {
	return r.message
}

func (r *RREvent) Plugin() string {
	return r.plugin
}
