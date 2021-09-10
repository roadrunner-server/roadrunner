# Migration from v1.0 to v2.0
To migration integration from RoadRunner v1.* to v2.* follow the next steps.

## Update Configuration
Second version of RoadRunner use single worker factory for all of its plugins. This means that you must include a new section
into your config `server` which is responsible for the worker creation. Limit service no longer presented as separate entity 
but rather part of specific service configuration.

```yaml
rpc:
  listen: tcp://127.0.0.1:6001

server:
  command: "php tests/psr-worker-bench.php"

http:
  address: "0.0.0.0:8080"
  pool:
    num_workers: 4
```

> Read more in [config reference](/intro/config.md).

## No longer worry about echoing
RoadRunner 2.0 intercepts all output to the STDOUT, this means you can start using default var_dump and other echo function
without breaking the communication. Yay!

## Explicitly declare PSR-15 dependency
We no longer ship the default PSR implementation with RoadRunner, make sure to include one you like the most by yourself.
For example:

```bash
$ composer require nyholm/psr7
```

## Update Worker Code
RoadRunner simplifies worker creation, use static `create()` method to automatically configure your worker:

```php
<?php

use Spiral\RoadRunner;

include "vendor/autoload.php";

$worker = RoadRunner\Worker::create();
```

Pass the PSR-15 factories to your PSR Worker:

```php
<?php

use Spiral\RoadRunner;
use Nyholm\Psr7;

include "vendor/autoload.php";

$worker = RoadRunner\Worker::create();
$psrFactory = new Psr7\Factory\Psr17Factory();

$worker = new RoadRunner\Http\PSR7Worker($worker, $psrFactory, $psrFactory, $psrFactory);
```

RoadRunner 2 unifies all workers to use similar naming, change `acceptRequest` to `waitRequest`:

```php
while ($req = $worker->waitRequest()) {
    try {
        $rsp = new Psr7\Response();
        $rsp->getBody()->write('Hello world!');

        $worker->respond($rsp);
    } catch (\Throwable $e) {
        $worker->getWorker()->error((string)$e);
    }
}
```

## Update RPCs
To create RPC client use new Goridge API:

```php
$rpc = \Spiral\Goridge\RPC\RPC::create('tcp://127.0.0.1:6001');
```