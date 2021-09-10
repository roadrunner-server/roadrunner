# Standalone Usage
You can use RoadRunner library as the component, without bringing the whole application into your project.

In order to create PHP worker pool on Golang:

```golang
srv := roadrunner.NewServer(
    &roadrunner.ServerConfig{
        Command: "php client.php echo pipes",
        Relay:   "pipes",
        Pool: &roadrunner.Config{
            NumWorkers:      int64(runtime.NumCPU()),
            AllocateTimeout: time.Second,
            DestroyTimeout:  time.Second,
        },
    })
defer srv.Stop()

srv.Start()

res, err := srv.Exec(&roadrunner.Payload{Body: []byte("hello")})
```

Worker (echo) structure would look like:

```php
<?php
/**
 * @var Goridge\RelayInterface $relay
 */
use Spiral\Goridge;
use Spiral\RoadRunner;

ini_set('display_errors', 'stderr');
require 'vendor/autoload.php';

$rr = new RoadRunner\Worker(new Spiral\Goridge\StreamRelay(STDIN, STDOUT));

while ($body = $rr->receive($context)) {
    try {
        $rr->send((string)$body, (string)$context);
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}
```

Make sure to run `go get github.com/spiral/roadrunner` and `composer require spiral/roadrunner` to load necessary dependencies.
