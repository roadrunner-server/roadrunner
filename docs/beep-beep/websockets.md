### Websockets

Websockets plugins add WebSockets broadcasting features to the Roadrunner. It
implements [PubSub](https://github.com/spiral/roadrunner/blob/master/pkg/pubsub/interface.go) interface.

#### Protobuf

Websockets plugin uses protobuf messages for the RPC calls (PHP part). The same messages, but JSON-encoded used on the
client side (browser, devices). All proto messages located in the
Roadrunner [pkg](https://github.com/spiral/roadrunner/tree/master/pkg/proto/websockets) folder.

#### RPC interface

1. `Publish(in *websocketsv1.Request, out *websocketsv1.Response)`: The arguments: first argument is a `Request` , which
   declares a `broker`, `topics` to push the payload and `payload`; the second argument is a `Response`, it will contain
   only 1 bool value which used as a signal of error.  
   The error returned if the request fails.

2. `PublishAsync(in *websocketsv1.Request, out *websocketsv1.Response)`: The arguments: first argument is a `Request` ,
   which declares a `broker`, `topics` to push the payload and `payload`; the second argument is a `Response`, it will
   contain only 1 bool value which used as a signal of error.  
   The difference between `Publish` and `PublishAsync` that `PublishAsync` doesn't wait for a broker's response. 
   The error returned if the request fails.

#### Clients

Client payload is the same as used in the RPC operations except that `command` field should be used. Commands can be as
following:

1. `join` - to join a specified topics. For successful `join` server returns a response with joined topics:
   `{"topic":"@join","payload":["foo","foo2"]}`. Otherwise, the server returns an error or unsuccessful `join` response:
   `{"topic":"#join","payload":["foo","foo2"]}`. 
   Sample of `join` command:`{"command":"join","broker":"memory","topics":["foo","foo2"],"payload":""}`


2. `leave` - to leave a specified topics. For successful `leave` server returns a response with a left topics:
   `{"topic":"@leave","payload":["foo","foo2"]}`. Otherwise, the server returns an error or unsuccessful `leave`
   response:
   `{"topic":"#leave","payload":["foo","foo2"]}`. Sample of `leave`
   command: `{"command":"leave","broker":"memory","topics":["foo","foo2"],"payload":""}`

#### Architecture

The architecture diagram for the WebSockets plugin can be
found [here](https://github.com/spiral/roadrunner/tree/master/plugins/websockets/doc)
