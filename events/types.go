package events

import (
	"fmt"
)

type EventBus interface {
	SubscribeAll(subID string, ch chan<- Event) error
	SubscribeP(subID string, pattern string, ch chan<- Event) error
	Unsubscribe(subID string)
	UnsubscribeP(subID, pattern string)
	Len() uint
	Send(ev Event)
}

type Event interface {
	Type() fmt.Stringer
	Plugin() string
	Message() string
}

type event struct {
	// event typ
	typ fmt.Stringer
	// plugin
	plugin string
	// message
	message string
}

// NewEvent initializes new event
func NewEvent(t fmt.Stringer, plugin string, message string) *event {
	if t.String() == "" || plugin == "" {
		return nil
	}

	return &event{
		typ:     t,
		plugin:  plugin,
		message: message,
	}
}

func (r *event) Type() fmt.Stringer {
	return r.typ
}

func (r *event) Message() string {
	return r.message
}

func (r *event) Plugin() string {
	return r.plugin
}
