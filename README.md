RoadRunner
==========
[![GoDoc](https://godoc.org/github.com/spiral/roadrunner?status.svg)](https://godoc.org/github.com/spiral/roadrunner)
[![Build Status](https://travis-ci.org/spiral/roadrunner.svg?branch=master)](https://travis-ci.org/spiral/roadrunner)

High-performance PHP job balancer for Golang.

Features:
--------
- load balancer, process manager and task pipeline 
- swaps workers without stopping the server
- build for multiple frontends (queue, rest, psr-7, async php, etc)
- works over TPC, unix sockets and standard pipes
- automatic worker replacement
- safe worker destruction
- worker lifecycle management (create/stop/allocate timeouts)
- payload context
- protocol, worker and job level error management
- very fast (~250k calls per second on Ryzen 1700X over 16 threads)
- works on Windows

Examples:
--------

```go
p, err := NewPool(
    func() *exec.Cmd { return exec.Command("php", "worker.php") },
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

while ($in = $rr->receive($ctx)) {
    try {
        $rr->send((string)$in);
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}
```

License:
--------
The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information.
