RoadRunner
==========
[![Latest Stable Version](https://poser.pugx.org/spiral/roadrunner/version)](https://packagist.org/packages/spiral/roadrunner)
[![GoDoc](https://godoc.org/github.com/spiral/roadrunner?status.svg)](https://godoc.org/github.com/spiral/roadrunner)
[![Build Status](https://travis-ci.org/spiral/roadrunner.svg?branch=master)](https://travis-ci.org/spiral/roadrunner)
[![Go Report Card](https://goreportcard.com/badge/github.com/spiral/roadrunner)](https://goreportcard.com/report/github.com/spiral/roadrunner)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/spiral/roadrunner/badges/quality-score.png)](https://scrutinizer-ci.com/g/spiral/roadrunner/?branch=master)
[![Codecov](https://codecov.io/gh/spiral/roadrunner/branch/master/graph/badge.svg)](https://codecov.io/gh/spiral/roadrunner/)

High-performance PHP load balancer, PSR-7 HTTP server and process manager.

Features:
--------
- PSR-7 HTTP server (file uploads, error handling, static files, hot reload, middlewares, event listeners)
- extendable service model
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
- very fast (~250k calls per second on Ryzen 1700X over 16 threads)
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
