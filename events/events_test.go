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
	err := eh.SubscribeP(id, "http.EventWorkerError", ch)
	require.NoError(t, err)

	eh.Send(NewEvent(EventWorkerError, "http", "foo"))

	evt := <-ch
	require.Equal(t, "foo", evt.Message())
	require.Equal(t, "http", evt.Plugin())
	require.Equal(t, "EventWorkerError", evt.Type().String())

	eh.Unsubscribe(id)
}

func TestEvenHandler2(t *testing.T) {
	eh, id := Bus()
	eh2, id2 := Bus()
	defer eh.Unsubscribe(id)
	defer eh2.Unsubscribe(id2)

	ch := make(chan Event, 100)
	ch2 := make(chan Event, 100)
	err := eh2.SubscribeP(id2, "http.EventWorkerError", ch)
	require.NoError(t, err)

	err = eh.SubscribeP(id, "http.EventWorkerError", ch2)
	require.NoError(t, err)

	eh.Send(NewEvent(EventWorkerError, "http", "foo"))

	evt := <-ch2
	require.Equal(t, "foo", evt.Message())
	require.Equal(t, "http", evt.Plugin())
	require.Equal(t, "EventWorkerError", evt.Type().String())

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

	eh.Unsubscribe(id)
	eh2.Unsubscribe(id2)
}

func TestEvenHandler3(t *testing.T) {
	eh, id := Bus()
	defer eh.Unsubscribe(id)

	ch := make(chan Event, 100)
	err := eh.SubscribeP(id, "EventWorkerError", ch)
	require.Error(t, err)

	eh.Unsubscribe(id)
}

func TestEvenHandler4(t *testing.T) {
	eh, id := Bus()
	defer eh.Unsubscribe(id)

	err := eh.SubscribeP(id, "EventWorkerError", nil)
	require.Error(t, err)

	eh.Unsubscribe(id)
}

func TestEvenHandler5(t *testing.T) {
	eh, id := Bus()
	defer eh.Unsubscribe(id)

	ch := make(chan Event, 100)
	err := eh.SubscribeP(id, "http.EventWorkerError", ch)
	require.NoError(t, err)

	eh.Send(NewEvent(EventWorkerError, "http", "foo"))

	evt := <-ch
	require.Equal(t, "foo", evt.Message())
	require.Equal(t, "http", evt.Plugin())
	require.Equal(t, "EventWorkerError", evt.Type().String())

	eh.Unsubscribe(id)
}

type MySuperEvent uint32

const (
	// EventHTTPError represents success unary call response
	EventHTTPError MySuperEvent = iota
)

func (mse MySuperEvent) String() string {
	switch mse {
	case EventHTTPError:
		return "EventHTTPError"
	default:
		return "UnknownEventType"
	}
}

func TestEvenHandler6(t *testing.T) {
	eh, id := Bus()
	defer eh.Unsubscribe(id)

	ch := make(chan Event, 100)
	err := eh.SubscribeP(id, "http.EventHTTPError", ch)
	require.NoError(t, err)

	eh.Send(NewEvent(EventHTTPError, "http", "foo"))

	evt := <-ch
	require.Equal(t, "foo", evt.Message())
	require.Equal(t, "http", evt.Plugin())
	require.Equal(t, "EventHTTPError", evt.Type().String())

	eh.Unsubscribe(id)
}

func TestEvenHandler7(t *testing.T) {
	eh, id := Bus()
	defer eh.Unsubscribe(id)

	ch := make(chan Event, 100)
	err := eh.SubscribeAll(id, ch)
	require.NoError(t, err)

	eh.Send(NewEvent(EventHTTPError, "http", "foo"))

	evt := <-ch
	require.Equal(t, "foo", evt.Message())
	require.Equal(t, "http", evt.Plugin())
	require.Equal(t, "EventHTTPError", evt.Type().String())

	eh.Unsubscribe(id)
}

func TestEvenHandler8(t *testing.T) {
	eh, id := Bus()
	defer eh.Unsubscribe(id)

	err := eh.SubscribeAll(id, nil)
	require.Error(t, err)

	eh.Unsubscribe(id)
}

func TestEvenHandler9(t *testing.T) {
	eh, id := Bus()
	defer eh.Unsubscribe(id)

	ch := make(chan Event, 100)
	err := eh.SubscribeP(id, "http.EventWorkerError", ch)
	require.NoError(t, err)

	eh.Send(NewEvent(EventWorkerError, "http", "foo"))

	evt := <-ch
	require.Equal(t, "foo", evt.Message())
	require.Equal(t, "http", evt.Plugin())
	require.Equal(t, "EventWorkerError", evt.Type().String())

	eh.UnsubscribeP(id, "http.EventWorkerError")

	eh.Send(NewEvent(EventWorkerError, "http", "foo"))

	select {
	case <-ch:
		require.Fail(t, "should not read any events")
	default:
		return
	}
}

func TestEvenHandler10(t *testing.T) {
	eh, id := Bus()
	defer eh.Unsubscribe(id)

	ch := make(chan Event, 100)
	err := eh.SubscribeP(id, "http.EventHTTPError", ch)
	require.NoError(t, err)
	err = eh.SubscribeP(id, "http.Foo", ch)
	require.NoError(t, err)
	err = eh.SubscribeP(id, "http.Foo2", ch)
	require.NoError(t, err)
	err = eh.SubscribeP(id, "http.Foo2", ch)
	require.NoError(t, err)

	eh.Send(NewEvent(EventHTTPError, "http", "foo"))

	evt := <-ch
	require.Equal(t, "foo", evt.Message())
	require.Equal(t, "http", evt.Plugin())
	require.Equal(t, "EventHTTPError", evt.Type().String())

	eh.Unsubscribe(id)
}
