# HTTP Middleware

RoadRunner HTTP server uses default Golang middleware model which allows you to extend it using custom or
community-driven middleware. The simplest service with middleware registration would look like:

```golang
package middleware

import (
 "net/http"
)

const PluginName = "middleware"

type Plugin struct{}

// to declare plugin
func (g *Plugin) Init() error {
    return nil
}

func (g *Plugin) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // do something
    // ...
    // continue request through the middleware pipeline
    next.ServeHTTP(w, r)
 })
}

// Middleware/plugin name.
func (g *Plugin) Name() string {
    return PluginName
}
```

> Middleware must correspond to the following [interface](https://github.com/spiral/roadrunner/blob/master/plugins/http/plugin.go#L37) and be named.

We have to register this service after in the [`internal/container/plugin.go`](https://github.com/spiral/roadrunner-binary/blob/master/internal/container/plugins.go#L31) file in order to properly resolve dependency:

```golang
import (
    "middleware"
)

func Plugins() []interface{} {
	return []interface{}{
    // ...
    
    // middleware
    &middleware.Plugin{},
    
    // ...
 }
```

You should also make sure you configure the middleware to be used via the [config or the command line](https://roadrunner.dev/docs/intro-config) otherwise the plugin will be loaded but the middleware will not be used with incoming requests.

```yaml
http:
    # provide the name of the plugin as provided by the plugin in the example's case, "middleware"
    middleware: [ "middleware" ]
```

### PSR7 Attributes

You can safely pass values to `ServerRequestInterface->getAttributes()` using [attributes](https://github.com/spiral/roadrunner/blob/master/plugins/http/attributes/attributes.go) package:

```golang
func (s *Service) middleware(next http.HandlerFunc) http.HandlerFunc {
 return func(w http.ResponseWriter, r *http.Request) {
     r = attributes.Init(r)
     attributes.Set(r, "key", "value")
            next(w, r)
 }
}
```
