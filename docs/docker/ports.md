# Ports and Containers
By default, embedded RPC server will listen to only localhost connections. In order to control RR from outside you must:

* Expose 6001 port from your container.
* Configure rr to listen on 0.0.0.0

```yaml
rpc:
  listen: tcp://:6001
```

> Remember that communication of TCP is slower than Unix sockets.
