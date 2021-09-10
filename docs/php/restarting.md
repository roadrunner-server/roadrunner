# Restarting Workers
RoadRunner provides multiple ways to safely restart worker(s) on demand. Both approaches can be used on a live server and should not cause downtime.

## Stop Command
You are able to send `stop` command from worker to parent server to force process destruction. In this scenario, 
the job/request will be automatically forwarded to the next worker.

We can demonstrate it by implementing `max_jobs` control on PHP end:

```php
<?php

use Spiral\RoadRunner;
use Nyholm\Psr7;

include "vendor/autoload.php";

$worker = RoadRunner\Worker::create();
$psrFactory = new Psr7\Factory\Psr17Factory();

$worker = new RoadRunner\Http\PSR7Worker($worker, $psrFactory, $psrFactory, $psrFactory);

$count = 0;
while ($req = $worker->waitRequest()) {
    try {
        $rsp = new Psr7\Response();
        $rsp->getBody()->write('Hello world!');

        $count++;
        if ($count > 10) {
            $worker->getWorker()->stop();
            return;
        }

        $worker->respond($rsp);
    } catch (\Throwable $e) {
        $worker->getWorker()->error((string)$e);
    }
}
```

> This approach can be used to control memory usage inside the PHP script.

## Full Reset
You can also initiate a rebuild of all RoadRunner workers using embedded [RPC bus](/beep-beep/rpc.md):

```php
$rpc = \Spiral\Goridge\RPC\RPC::create('tcp://127.0.0.1:6001');
$rpc->call('resetter.Reset', 'http');
```
