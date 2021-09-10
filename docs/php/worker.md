# PHP Workers

In order to run your PHP application, you must create a worker endpoint and configure RoadRunner to use it. First,
install the required package using [Composer](https://getcomposer.org/).

```bash
composer require spiral/roadrunner nyholm/psr7
```

Simplest entrypoint with PSR-7 server API might looks like:

```php
<?php

use Spiral\RoadRunner;
use Nyholm\Psr7;

include "vendor/autoload.php";

$worker = RoadRunner\Worker::create();
$psrFactory = new Psr7\Factory\Psr17Factory();

$worker = new RoadRunner\Http\PSR7Worker($worker, $psrFactory, $psrFactory, $psrFactory);

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

Such a worker will expect communication with the parent RoadRunner server over standard pipes, create `.rr.yaml` config
to enable it:

```yaml
server:
  command: "php psr-worker.php"

http:
  address: 0.0.0.0:8080
  pool:
    num_workers: 4
```

If you don't like `yaml` try `.rr.json`:

```json
{
  "server": {
    "command": "path-to-php/php psr-worker.php"
  },
  "http": {
    "address": "0.0.0.0:8080",
    "pool": {
      "num_workers": 4
    }
  }
}
```

You can start the application now by downloading the RR binary file and running `rr serve`

## Alternative Communication Methods

PHP Workers would utilize standard pipes STDOUT and STDERR to exchange data frames with RR server. In some cases you might
want to use alternative communication methods such as TCP socket:

```yaml
server:
  command: "php psr-worker.php"
  relay: "tcp://localhost:7000"
  
http:
  address: 0.0.0.0:8080
  pool:
    num_workers: 4
```

Unix sockets:

```yaml
server:
  command: "php psr-worker.php"
  relay:  "unix://rr.sock"

http:
  address: 0.0.0.0:8080
  pool:
    num_workers: 4
```

## Troubleshooting

In some cases, RR would not be able to handle errors produced by PHP worker (PHP is missing, the script is dead etc)
.

```
$ rr serve
DEBU[0003] [rpc]: started
DEBU[0003] [http]: started
ERRO[0003] [http]: unable to connect to worker: unexpected response, header is missing: exit status 1
DEBU[0003] [rpc]: stopped
```

You can troubleshoot it by invoking `command` set in `.rr` file manually:

```
$ php psr-worker.php
```

The worker should not cause any error until any input provided and must fail with invalid input signature after the
first input character.

## Other Type of Workers

Different roadrunner implementations might define their own worker APIs,
examples: [GRPC](https://github.com/spiral/php-grpc), [Workflow/Activity Worker](https://docs.temporal.io/docs/php-workers/).