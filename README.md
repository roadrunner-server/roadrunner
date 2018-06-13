RoadRunner
==========
[![Latest Stable Version](https://poser.pugx.org/spiral/roadrunner/version)](https://packagist.org/packages/spiral/roadrunner)
[![GoDoc](https://godoc.org/github.com/spiral/roadrunner?status.svg)](https://godoc.org/github.com/spiral/roadrunner)
[![Build Status](https://travis-ci.org/spiral/roadrunner.svg?branch=master)](https://travis-ci.org/spiral/roadrunner)
[![Go Report Card](https://goreportcard.com/badge/github.com/spiral/roadrunner)](https://goreportcard.com/report/github.com/spiral/roadrunner)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/spiral/roadrunner/badges/quality-score.png)](https://scrutinizer-ci.com/g/spiral/roadrunner/?branch=master)
[![Codecov](https://codecov.io/gh/spiral/roadrunner/branch/master/graph/badge.svg)](https://codecov.io/gh/spiral/roadrunner/)

High-performance PSR-7 HTTP server, PHP load balancer and process manager.

Features:
--------
- PSR-7 HTTP server (file uploads, error handling, static files, hot reload, middlewares, event listeners)
- extendable service model (plus PHP compatible RPC server)
- no external services, drop-in (based on [Goridge](https://github.com/spiral/goridge))
- load balancer, process manager and task pipeline
- frontend agnostic (queue, REST, PSR-7, async php, etc)
- works over TCP, unix sockets and standard pipes
- automatic worker replacement and safe PHP process destruction
- worker lifecycle management (create/allocate/destroy timeouts)
- payload context and body
- control over max jobs per worker
- protocol, worker and job level error management (including PHP errors)
- memory leak failswitch
- very fast (~250k rpc calls per second on Ryzen 1700X over 16 threads)
- works on Windows

Installation:
--------
```
$ go get github.com/spiral/roadrunner
$ composer require spiral/roadrunner
```

Usage:
------

```
$ cd cmd
$ cd rr
$ go build && go install
$ cp .rr.yaml path/to/the/project
```

> TODO: To be updated with build scripts!

```
$ rr serve -v
```

Example [worker](https://github.com/spiral/roadrunner/blob/master/php-src/tests/http/client.php).

Example config: 
---------------

```yaml
# rpc bus allows php application and external clients to talk to rr services.
rpc:
  # enable rpc server
  enable: true

  # rpc connection DSN. Supported TCP and Unix sockets.
  listen:     tcp://127.0.0.1:6001

# http service configuration.
http:
  # set to false to disable http server.
  enable:    true

  # http host to listen.
  address:   0.0.0.0:8080

  # max POST request size, including file uploads in MB.
  maxRequest: 200

  # file upload configuration.
  uploads:
    # list of file extensions which are forbidden for uploading.
    forbid: [".php", ".exe", ".bat"]

  # http worker pool configuration.
  workers:
    # php worker command.
    command:  "php psr-worker.php pipes"

    # connection method (pipes, tcp://:9000, unix://socket.unix).
    relay:    "pipes"

    # worker pool configuration.
    pool:
      # number of workers to be serving.
      numWorkers: 4

      # maximum jobs per worker, 0 - unlimited.
      maxJobs:  0

      # for how long pool should attempt to allocate free worker (request timeout). In nanoseconds for now :(
      allocateTimeout: 600000000

      # amount of time given to worker to gracefully destruct itself. In nanoseconds for now :(
      destroyTimeout:  600000000

# static file serving.
static:
  # serve http static files
  enable:  false

  # root directory for static file (http would not serve .php and .htaccess files).
  dir:   "public"

  # list of extensions to forbid for serving.
  forbid: [".php", ".htaccess"]
```

Examples:
--------

```go
p, err := rr.NewPool(
    func() *exec.Cmd { return exec.Command("php", "worker.php", "pipes") },
    rr.NewPipeFactory(),
    rr.Config{
        NumWorkers:      uint64(runtime.NumCPU()),
        AllocateTimeout: time.Second,              
        DestroyTimeout:  time.Second,               
    },
)
defer p.Destroy()

rsp, err := p.Exec(&rr.Payload{Body: []byte("hello")})
```
```php
<?php
/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;

$rr = new RoadRunner\Worker($relay);

while ($body = $rr->receive($context)) {
    try {
        $rr->send((string)$body, (string)$context);
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}
```
> Check how to init relay [here](./php-src/tests/client.php). More examples can be found in tests.

Testing:
--------
```
$ make test
```

License:
--------
The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information.
