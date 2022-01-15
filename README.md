<p align="center">
 <img src="https://user-images.githubusercontent.com/796136/50286124-6f7f3780-046f-11e9-9f45-e8fedd4f786d.png" height="75px" alt="RoadRunner">
</p>
<p align="center">
 <a href="https://github.com/spiral/roadrunner-binary/releases"><img src="https://img.shields.io/github/v/release/spiral/roadrunner-binary.svg?maxAge=30"></a>
	<a href="https://pkg.go.dev/github.com/spiral/roadrunner-binary/v2"><img src="https://godoc.org/github.com/spiral/roadrunner-binary/v2?status.svg"></a>
	<a href="https://github.com/spiral/roadrunner-binary/actions"><img src="https://github.com/spiral/roadrunner-binary/workflows/tests/badge.svg"></a>
	<a href="https://goreportcard.com/report/github.com/spiral/roadrunner-binary"><img src="https://goreportcard.com/badge/github.com/spiral/roadrunner-binary"></a>
	<a href="https://lgtm.com/projects/g/spiral/roadrunner-binary/alerts/"><img alt="Total alerts" src="https://img.shields.io/lgtm/alerts/g/spiral/roadrunner-binary.svg?logo=lgtm&logoWidth=18"/></a>
	<a href="https://discord.gg/TFeEmCs"><img src="https://img.shields.io/badge/discord-chat-magenta.svg"></a>
    <img alt="All releases" src="https://img.shields.io/github/downloads/spiral/roadrunner-binary/total">
</p>

RoadRunner is an open-source (MIT licensed) high-performance PHP application server, load balancer, and process manager.
It supports running as a service with the ability to extend its functionality on a per-project basis.

RoadRunner includes PSR-7/PSR-17 compatible HTTP and HTTP/2 server and can be used to replace classic Nginx+FPM setup
with much greater performance and flexibility.

<p align="center">
	<a href="https://roadrunner.dev/"><b>Official Website</b></a> |
	<a href="https://roadrunner.dev/docs"><b>Documentation</b></a> |
    <a href="https://github.com/orgs/spiral/projects/2"><b>Release schedule</b></a>
</p>

Features:
--------
- Production-ready
- PCI DSS compliant
- PSR-7 HTTP server (file uploads, error handling, static files, hot reload, middlewares, event listeners)
- HTTPS and HTTP/2 support (including HTTP/2 Push, H2C)
- A Fully customizable server, FastCGI support
- Flexible environment configuration
- No external PHP dependencies (64bit version required), drop-in (based on [Goridge](https://github.com/spiral/goridge))
- Load balancer, process manager and task pipeline
- Integrated metrics (Prometheus)
- [Workflow engine](https://github.com/temporalio/sdk-php) by [Temporal.io](https://temporal.io)
- Works over TCP, UNIX sockets and standard pipes
- Automatic worker replacement and safe PHP process destruction
- Worker create/allocate/destroy timeouts
- Max jobs per worker
- Worker lifecycle management (controller)
  - maxMemory (graceful stop)
  - TTL (graceful stop)
  - idleTTL (graceful stop)
  - execTTL (brute, max_execution_time)
- Payload context and body
- Protocol, worker and job level error management (including PHP errors)
- Development Mode
- Integrations with Symfony, [Laravel](https://github.com/spiral/roadrunner-laravel), Slim, CakePHP, Zend Expressive
- Application server for [Spiral](https://github.com/spiral/framework)
- Included in Laravel Octane
- Automatic reloading on file changes
- Works on Windows (Unix sockets (AF_UNIX) supported on Windows 10)

Installation:
--------

```bash
$ composer require spiral/roadrunner:v2.0 nyholm/psr7
$ ./vendor/bin/rr get-binary
```

> For getting roadrunner binary file you can use our docker image: `spiralscout/roadrunner:X.X.X` (more information about
> image and tags can be found [here](https://hub.docker.com/r/spiralscout/roadrunner/)).
>
> Important notice: we strongly recommend to use a versioned tag (like `1.2.3`) instead `latest`.

Configuration can be located in `.rr.yaml`
file ([full sample](https://github.com/spiral/roadrunner/blob/master/.rr.yaml)):

```yaml
rpc:
  listen: tcp://127.0.0.1:6001

server:
  command: "php worker.php"

http:
  address: "0.0.0.0:8080"

logs:
  level: error
```

> Read more in [Documentation](https://roadrunner.dev/docs).

Example Worker:
--------

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

Run:
----
To run application server:

```
$ ./rr serve
```

License:
--------
The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information. Maintained
by [Spiral Scout](https://spiralscout.com).
