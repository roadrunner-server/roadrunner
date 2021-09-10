# Headers and CORS
RoadRunner can automatically set up request/response headers and control CORS for your application.

### CORS
To enable CORS headers add the following section to your configuration.

```yaml
http:
  address: 127.0.0.1:44933
  middleware: ["headers"]
  # ...
  headers:
    cors:
      allowed_origin: "*"
      allowed_headers: "*"
      allowed_methods: "GET,POST,PUT,DELETE"
      allow_credentials: true
      exposed_headers: "Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma"
      max_age: 600
```

> Make sure to declare "headers" middleware.

### Custom headers for Response or Request
You can control additional headers to be set for outgoing responses and headers to be added to the request sent to your application.
```yaml
http:
  # ...
  headers:
      # Automatically add headers to every request passed to PHP.
      request:
        Example-Request-Header: "Value"
    
      # Automatically add headers to every response.
      response:
        X-Powered-By: "RoadRunner"
```
