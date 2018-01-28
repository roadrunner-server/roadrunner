RoadRunner
==========
[![GoDoc](https://godoc.org/github.com/spiral/roadrunner?status.svg)](https://godoc.org/github.com/spiral/roadrunner)
[![Build Status](https://travis-ci.org/spiral/roadrunner.svg?branch=master)](https://travis-ci.org/spiral/roadrunner)
[![Go Report Card](https://goreportcard.com/badge/github.com/spiral/roadrunner)](https://goreportcard.com/report/github.com/spiral/roadrunner)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/spiral/roadrunner/badges/quality-score.png)](https://scrutinizer-ci.com/g/spiral/roadrunner/?branch=master)
[![Codecov](https://codecov.io/gh/spiral/roadrunner/branch/master/graph/badge.svg)](https://codecov.io/gh/spiral/roadrunner/)

High-performance PHP job balancer and process manager library for Golang.

Features:
--------
- no external dependencies or services, drop-in
- load balancer, process manager and task pipeline
- build for multiple frontends (queue, rest, psr-7, async php, etc)
- works over TPC, unix sockets and standard pipes
- automatic worker replacement and safe destruction
- worker lifecycle management (create/allocate/destroy timeouts)
- payload context and body
- control over max jobs per worker
- protocol, worker and job level error management (including PHP errors)
- very fast (~250k calls per second on Ryzen 1700X over 16 threads)
- works on Windows

Examples:
--------

```go
p, err := NewPool(
    func() *exec.Cmd { return exec.Command("php", "worker.php", "pipes") },
    NewPipeFactory(),
    Config{
        NumWorkers:      uint64(runtime.NumCPU()),
        AllocateTimeout: time.Second,              
        DestroyTimeout:  time.Second,               
    },
)
defer p.Destroy()

rsp, err := p.Exec(&Payload{Body: []byte("hello")})
```
```php
<?php
/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;

$rr = new RoadRunner\Worker($relay);

while ($in = $rr->receive($context)) {
    try {
        $rr->send((string)$in, (string)$context);
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}
```
> Check how to init relay [here](./tests/client.php).

License:
--------
The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information.
