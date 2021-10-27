package events

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEvenHandler(t *testing.T) {
	eh, id := Bus()
	defer eh.Unsubscribe(id)

	ch := make(chan Event, 100)
	err := eh.SubscribeP(id, "http.EventJobOK", ch)
	require.NoError(t, err)

	eh.Send(NewEvent(EventJobOK, "http", "foo"))

	evt := <-ch
	require.Equal(t, "foo", evt.Message())
	require.Equal(t, "http", evt.Plugin())
	require.Equal(t, "EventJobOK", evt.Type().String())
}

func TestEvenHandler2(t *testing.T) {
	eh, id := Bus()
	eh2, id2 := Bus()
	defer eh.Unsubscribe(id)
	defer eh2.Unsubscribe(id2)

	ch := make(chan Event, 100)
	ch2 := make(chan Event, 100)
	err := eh2.SubscribeP(id2, "http.EventJobOK", ch)
	require.NoError(t, err)

	err = eh.SubscribeP(id, "http.EventJobOK", ch2)
	require.NoError(t, err)

	eh.Send(NewEvent(EventJobOK, "http", "foo"))

	evt := <-ch2
	require.Equal(t, "foo", evt.Message())
	require.Equal(t, "http", evt.Plugin())
	require.Equal(t, "EventJobOK", evt.Type().String())

	l := eh.Len()
	require.Equal(t, uint(2), l)

	eh.Unsubscribe(id)
	time.Sleep(time.Second)

	l = eh.Len()
	require.Equal(t, uint(1), l)

	eh2.Unsubscribe(id2)
	time.Sleep(time.Second)

	l = eh.Len()
	require.Equal(t, uint(0), l)
}

func TestEvenHandler3(t *testing.T) {
	eh, id := Bus()
	defer eh.Unsubscribe(id)

	ch := make(chan Event, 100)
	err := eh.SubscribeP(id, "EventJobOK", ch)
	require.Error(t, err)
}

func TestEvenHandler4(t *testing.T) {
	eh, id := Bus()
	defer eh.Unsubscribe(id)

	err := eh.SubscribeP(id, "EventJobOK", nil)
	require.Error(t, err)
}

func TestEvenHandler5(t *testing.T) {
	eh, id := Bus()
	defer eh.Unsubscribe(id)

	ch := make(chan Event, 100)
	err := eh.SubscribeP(id, "http.EventJobOK", ch)
	require.NoError(t, err)

	eh.Send(NewEvent(EventJobOK, "http", "foo"))

	evt := <-ch
	require.Equal(t, "foo", evt.Message())
	require.Equal(t, "http", evt.Plugin())
	require.Equal(t, "EventJobOK", evt.Type().String())
}
