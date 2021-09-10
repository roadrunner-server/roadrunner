# RPC Integration

You can connect to RoadRunner server from your PHP workers using shared RPC bus. In order to do that you have to create
an instance of `RPC` class configured to work with the address specified in `.rr` file.

## Requirements

To connect to RoadRunner from PHP application in RPC mode you need:

- ext-sockets
- ext-json

## Configuration

To change the RPC port from the default (localhost:6001) use:

```yaml
rpc:
  listen: tcp://127.0.0.1:6001
```

```php
$rpc = Goridge\RPC\RPC::create(RoadRunner\Environment::fromGlobals()->getRPCAddress());
```

You can immediately use this RPC to call embedded RPC services such as HTTP:

```php
var_dump($rpc->call('informer.Workers', 'http'));
```

You can read how to create your own services and RPC methods in [this section](/beep-beep/plugin.md).
