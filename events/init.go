package events

import (
	"sync"

	"github.com/google/uuid"
)

var evBus *eventsBus
var onInit = &sync.Once{}

func Bus() (*eventsBus, string) {
	onInit.Do(func() {
		evBus = newEventsBus()
		go evBus.handleEvents()
	})

	// return events bus with subscriberID
	return evBus, uuid.NewString()
}
