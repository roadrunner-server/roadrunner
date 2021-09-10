# RPC to App Server
You can connect to application server via `SocketRelay`:

```php
$rpc = \Spiral\Goridge\RPC\RPC::create('tcp://127.0.0.1:6001');
```

You can immediately use this RPC to call embedded RPC services such as HTTP:

```php
var_dump($rpc->call('informer.Workers', 'http'));
```

> Please note that in the case of running workers in debug mode (`http: { debug: true }` in `.rr.yaml`) the number 
> of http workers will be zero (i.e. an empty array `[]` will be returned).
> 
> This behavior may be changed in the future, you should not rely on this result to check that the 
> RoadRunner was launched in development mode.

You can read how to create your own services and RPC methods in [this section](/beep-beep/plugin.md).
