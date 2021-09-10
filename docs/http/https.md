# HTTPS and HTTP/2

You can enable HTTPS and HTTP2 support by adding `ssl` section into `http` config.

```yaml
http:
  # host and port separated by semicolon
  address: 127.0.0.1:8080
 
  ssl:
    # host and port separated by semicolon (default :443)
    address: :8892
    redirect: false
    cert: fixtures/server.crt
    key: fixtures/server.key
    root_ca: root.crt
  
  # optional support for http2  
  http2:
    h2c: false
    max_concurrent_streams: 128
```

### Redirecting HTTP to HTTPS

To enable an automatic redirect from `http://` to `https://` set `redirect` option to `true` (disabled by default).

### HTTP/2 Push Resources

RoadRunner support [HTTP/2 push](https://en.wikipedia.org/wiki/HTTP/2_Server_Push) via virtual headers provided by PHP
response.

```php
return $response->withAddedHeader('http2-push', '/test.js');
```

Note that the path of the resource must be related to the public application directory and must include `/` at the
beginning.

> Please note, HTTP2 push only works under HTTPS with `static` service enabled.

## H2C

You can enable HTTP/2 support over non-encrypted TCP connection using H2C:

```yaml
http:
  http2.h2c: true
```

### FastCGI

There is FastCGI frontend support inside the HTTP module, you can enable it (disabled by default):

```yaml
http:
  # HTTP service provides FastCGI as frontend
  fcgi:
    # FastCGI connection DSN. Supported TCP and Unix sockets.
    address: tcp://0.0.0.0:6920
```

### Root certificate authority support

Root CA supported by the option in .rr.yaml

```yaml
http:
  ssl:
    root_ca: root.crt
```