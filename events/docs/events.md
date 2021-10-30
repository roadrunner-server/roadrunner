## RoadRunner Events bus

RR events bus might be useful when one plugin raises some event on which another plugin should react. For example,
plugins like sentry might log errors from the `http` plugin.

## Simple subscription

Events bus supports wildcard subscriptions on the events as well as the direct subscriptions on the particular event.

Let's have a look at the simple example:

```go
package foo

import (
    "github.com/spiral/roadrunner/v2/events"
)

func foo() {
    eh, id := events.Bus()
    defer eh.Unsubscribe(id)

    ch := make(chan events.Event, 100)
    err := eh.SubscribeP(id, "http.EventJobOK", ch)
    if err != nil {
        panic(err)
    }

    eh.Send(events.NewEvent(events.EventJobOK, "http", "foo"))
    evt := <-ch
    // evt.Message() -> "foo"
    // evt.Plugin() -> "http"
    // evt.Type().String() -> "EventJobOK"
}
```

Here:
1. `eh, id := events.Bus()` get the instance (it's global) of the events bus. Make sure to unsubscribe event handler when you don't need it anymore: `eh.Unsubscribe(id)`.
2. `ch := make(chan events.Event, 100)` create an events channel.
3. `err := eh.SubscribeP(id, "http.EventJobOK", ch)` subscribe to the events which fits your pattern (`http.EventJobOK`).
4. `eh.Send(events.NewEvent(events.EventJobOK, "http", "foo"))` emit event from the any plugin.
5. `evt := <-ch` get the event.

Notes:
1. If you use only `eh.Send` events bus function, you don't need to unsubscribe, so, you may simplify the declaration to the `eh, _ := events.Bus()`.

## Wildcards

As mentioned before, RR events bus supports wildcards subscriptions, like: `*.SomeEvent`, `http.*`, `http.Some*`, `*`.
Let's have a look at the next sample of code:

```go
package foo

import (
    "github.com/spiral/roadrunner/v2/events"
)

func foo() {
    eh, id := events.Bus()
    defer eh.Unsubscribe(id)

    ch := make(chan events.Event, 100)
    err := eh.SubscribeP(id, "http.*", ch)
    if err != nil {
        panic(err)
    }

    eh.Send(events.NewEvent(events.EventJobOK, "http", "foo"))
    evt := <-ch
    // evt.Message() -> "foo"
    // evt.Plugin() -> "http"
    // evt.Type().String() -> "EventJobOK"
}
```

One change between these samples is in the `SubscribeP` pattern: `err := eh.SubscribeP(id, "http.*", ch)`. Here we used `http.*` instead of `http.EventJobOK`.

You also have the access to the event message, plugin and type. Message is a custom, user-defined message to log or to show to the subscriber. Plugin is a source plugin who raised this event. And the event type - is your custom or RR event type.


## How to implement custom event

Event type is a `fmt.Stringer`. That means, that your custom event type should implement this interface. Let's have a look at how to do that:

```go
package foo

type MySuperEvent uint32

const (
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
```

Here we defined a custom type - `MySuperEvent`. For sure, it might have any name you want and represent for example some domain field like `WorkersPoolEvent` represents RR sync.pool events. Then you need to implement a `fmt.Stringer` on your custom event type.
Next you need to create an enum with the actual events and that's it.

How to use that:
```go
package foo

import (
    "github.com/spiral/roadrunner/v2/events"
)

func foo() {
    eh, id := events.Bus()
    defer eh.Unsubscribe(id)

    ch := make(chan events.Event, 100)
    err := eh.SubscribeP(id, "http.EventHTTPError", ch)
    if err != nil {
        panic(err)
    }

	// first arg of the NewEvent method is fmt.Stringer
    eh.Send(events.NewEvent(EventHTTPError, "http", "foo"))
    evt := <-ch
    // evt.Message() -> "foo"
    // evt.Plugin() -> "http"
    // evt.Type().String() -> "EventHTTPError"
}
```

Important note: you don't need to import your custom event types into the subscriber. You only need to know the name of that event and pass a string to the subscriber.

