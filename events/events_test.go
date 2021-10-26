package events

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvenHandler(t *testing.T) {
	eh, id := Bus()

	ch := make(chan Event, 100)
	err := eh.SubscribeP(id, "http.EventJobOK", ch)
	require.NoError(t, err)

	eh.Send(NewRREvent(EventJobOK, "foo", "http"))

	evt := <-ch
	require.Equal(t, "foo", evt.Message())
	require.Equal(t, "http", evt.Plugin())
	require.Equal(t, "EventJobOK", evt.Type().String())
}
