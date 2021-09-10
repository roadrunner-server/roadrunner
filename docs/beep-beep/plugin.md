# Writing Plugins

RoadRunner uses Endure container to manage dependencies. This approach is similar to the PHP Container implementation
with automatic method injection. You can create your own plugins, event listeners, middlewares, etc.

To define your plugin, create a struct with public `Init` method with error return value (you can use `spiral/errors` as
the `error` package):

```golang
package custom

const PluginName = "custom"

type Plugin struct{}

func (s *Plugin) Init() error {
     return nil
}
```

You can register your plugin by creating a custom version of `main.go` file and [building it](/beep-beep/build.md).

### Dependencies

You can access other RoadRunner plugins by requesting dependencies in your `Init` method:

```golang
package custom

import (
     "github.com/spiral/roadrunner/v2/plugins/http"
     "github.com/spiral/roadrunner/v2/plugins/rpc"
)

type Service struct {}

func (s *Service) Init(r *rpc.Plugin, rr *http.Plugin) error {
     return nil
}
```

> Make sure to request dependency as pointer.

### Configuration

In most of the cases, your services would require a set of configuration values. RoadRunner can automatically populate
and validate your configuration structure using `config` plugin (via an interface):

Config sample:

```yaml
custom:
  address: tcp://localhost:8888
```

Plugin:

```golang
package custom

import (
     "github.com/spiral/roadrunner/v2/plugins/config"
     "github.com/spiral/roadrunner/v2/plugins/http"
     "github.com/spiral/roadrunner/v2/plugins/rpc"

     "github.com/spiral/errors"
)

const PluginName = "custom"

type Config struct{
     Address string `mapstructure:"address"`
}

type Plugin struct {
     cfg *Config
}

// You can also initialize some defaults values for config keys
func (cfg *Config) InitDefaults() {
     if cfg.Address == "" {
      cfg.Address = "tcp://localhost:8088"
    }
}

func (s *Plugin) Init(r *rpc.Plugin, h *http.Plugin, cfg config.Configurer) error {
 const op = errors.Op("custom_plugin_init") // error operation name
 if !cfg.Has(PluginName) {
  return errors.E(op, errors.Disabled)
 }

 // unmarshall 
 err := cfg.UnmarshalKey(PluginName, &s.cfg)
 if err != nil {
  // Error will stop execution
  return errors.E(op, err)
 }

 // Init defaults
 s.cfg.InitDefaults()
 
 return nil
}

```

`errors.Disabled` is the special kind of error which indicated Endure to disable this plugin and all dependencies of
this root. The RR2 will continue to work after this error type if at least plugin stay alive.

### Serving

Create `Serve` and `Stop` method in your structure to let RoadRunner start and stop your service.

```golang
type Plugin struct {}

func (s *Plugin) Serve() chan error {
 const op = errors.Op("custom_plugin_serve")
    errCh := make(chan error, 1)
    
    err := s.DoSomeWork()
    err != nil {
     errCh <- errors.E(op, err)
     return errCh
    }
    
    return nil
}

func (s *Plugin) Stop() error {
    return s.stopServing()
}

func (s *Plugin) DoSomeWork() error {
 return nil
}
```

`Serve` method is thread-safe. It runs in the separate goroutine which managed by the `Endure` container. The one note, is that you should unblock it when call `Stop` on the container. Otherwise, service will be killed after timeout (can be set in Endure).

### Collecting dependencies in runtime

RR2 provide a way to collect dependencies in runtime via `Collects` interface. This is very useful for the middlewares or extending plugins with additional functionality w/o changing it.
Let's create an HTTP middleware:

Steps (sample based on the actual `http` plugin and `Middleware` interface):

1. Declare a required interface

```go
// Middleware interface
type Middleware interface {
 Middleware(f http.Handler) http.HandlerFunc
}
```

2. Implement method, which should have as an arguments name (`endure.Named` interface) and `Middleware` (step 1).

```go
// Collects collecting http middlewares
func (s *Plugin) AddMiddleware(name endure.Named, m Middleware) {
    s.mdwr[name.Name()] = m
}
```

3. Implement `Collects` endure interface for the required structure and return implemented on the step 2 method.

```golang
// Collects collecting http middlewares
func (s *Plugin) Collects() []interface{} {
    return []interface{}{
        s.AddMiddleware,
    }
}
```

Endure will automatically check that registered structure implement all the arguments for the `AddMiddleware` method (or will find a structure if argument is structure). In our case, a structure should implement `endure.Named` interface (which returns user friendly name for the plugin) and `Middleware` interface.

### RPC Methods

You can expose a set of RPC methods for your PHP workers also by using Endure `Collects` interface. Endure will automatically get the structure and expose RPC method under the `PluginName` name.

To extend your plugin with RPC methods, plugin will not be changed at all. Only 1 thing to do is to create a file with RPC methods (let's call it `rpc.go`) and expose here all RPC methods for the plugin w/o changing plugin itself:
Sample based on the `informer` plugin:

I assume we created a file `rpc.go`. The next step is to create a structure:

1. Create a structure: (logger is optional)

```golang
package custom

type rpc struct {
 srv *Plugin
 log logger.Logger
}
```

2. Create a method, which you want to expose:

```go
func (s *rpc) Hello(input string, output *string) error {
 *output = input
 return nil
}
```

3. Use `Collects` interface to expose the RPC service to Endure:

```go
// CollectTarget resettable service.
func (p *Plugin) CollectTarget(name endure.Named, r Informer) error {
 p.registry[name.Name()] = r
 return nil
}

// Collects declares services to be collected.
func (p *Plugin) Collects() []interface{} {
 return []interface{}{
  p.CollectTarget,
 }
}

// Name of the service.
func (p *Plugin) Name() string {
 return PluginName
}

// RPCService returns associated rpc service.
func (p *Plugin) RPC() interface{} {
 return &rpc{srv: p, log: p.log}
}
```

Let's take a look at these methods:

1. `CollectTarget`: tells Endure, that we want to collect all plugins which implement `endure.Named` and `Informer` interfaces.
2. `Collects`: Endure interface implementation.
3. `Name`: `endure.Named` interface implementation which return a user-friendly plugin name.
4. `RPC`: RPC plugin Collects all plugins which implement `RPC` interface and `endure.Named`. RPC interface accepts no arguments, but returns interface (plugin).

To use it within PHP using `RPC` [instance](/beep-beep/rpc.md):

```php
var_dump($rpc->call('custom.Hello', 'world'));
```
