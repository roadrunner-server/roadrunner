package events

import (
	"fmt"
	"sync"
)

type sub struct {
	pattern string
	w       *wildcard
	events  chan<- Event
}

type eventsBus struct {
	sync.RWMutex
	subscribers  sync.Map
	internalEvCh chan Event
	stop         chan struct{}
}

func newEventsBus() *eventsBus {
	return &eventsBus{
		internalEvCh: make(chan Event, 100),
		stop:         make(chan struct{}),
	}
}

/*
http.* <-
*/

// SubscribeAll for all RR events
// returns subscriptionID
func (eb *eventsBus) SubscribeAll(subID string, ch chan<- Event) error {
	return eb.subscribe(subID, "*", ch)
}

// SubscribeP pattern like "pluginName.EventType"
func (eb *eventsBus) SubscribeP(subID string, pattern string, ch chan<- Event) error {
	return eb.subscribe(subID, pattern, ch)
}

func (eb *eventsBus) Unsubscribe(subID string) {
	eb.subscribers.Delete(subID)
}
func (eb *eventsBus) UnsubscribeP(subID, pattern string) {
	if sb, ok := eb.subscribers.Load(subID); ok {
		eb.Lock()
		defer eb.Unlock()

		sbArr := sb.([]*sub)

		for i := 0; i < len(sbArr); i++ {
			if sbArr[i].pattern == pattern {
				sbArr[i] = sbArr[len(sbArr)-1]
				sbArr = sbArr[:len(sbArr)-1]
				// replace with new array
				eb.subscribers.Store(subID, sbArr)
				return
			}
		}
	}
}

// Send sends event to the events bus
func (eb *eventsBus) Send(ev Event) {
	eb.internalEvCh <- ev
}

func (eb *eventsBus) subscribe(subID string, pattern string, ch chan<- Event) error {
	eb.Lock()
	defer eb.Unlock()
	w, err := newWildcard(pattern)
	if err != nil {
		return err
	}

	if sb, ok := eb.subscribers.Load(subID); ok {
		// at this point we are confident that sb is a []*sub type
		subArr := sb.([]*sub)
		subArr = append(subArr, &sub{
			pattern: pattern,
			w:       w,
			events:  ch,
		})

		eb.subscribers.Store(subID, subArr)

		return nil
	}

	subArr := make([]*sub, 0, 5)
	subArr = append(subArr, &sub{
		pattern: pattern,
		w:       w,
		events:  ch,
	})

	eb.subscribers.Store(subID, subArr)

	return nil
}

func (eb *eventsBus) handleEvents() {
	for { //nolint:gosimple
		select {
		case ev := <-eb.internalEvCh:
			//
			wc := fmt.Sprintf("%s.%s", ev.Plugin(), ev.Type().String())

			eb.subscribers.Range(func(key, value interface{}) bool {
				vsub := value.([]*sub)

				for i := 0; i < len(vsub); i++ {
					if vsub[i].w.match(wc) {
						select {
						case vsub[i].events <- ev:
							return true
						default:
							return true
						}
					}
				}

				return true
			})
		}
	}
}
