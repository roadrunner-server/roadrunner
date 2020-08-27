<p align="center">
 <img src="https://user-images.githubusercontent.com/796136/50286124-6f7f3780-046f-11e9-9f45-e8fedd4f786d.png" height="75px" alt="RoadRunner">
</p>
<p align="center">
 <a href="https://packagist.org/packages/spiral/roadrunner"><img src="https://poser.pugx.org/spiral/roadrunner/version"></a>
	<a href="https://pkg.go.dev/github.com/spiral/roadrunner?tab=doc"><img src="https://godoc.org/github.com/spiral/roadrunner?status.svg"></a>
	<a href="https://github.com/spiral/roadrunner/actions"><img src="https://github.com/spiral/roadrunner/workflows/CI/badge.svg" alt=""></a>
	<a href="https://goreportcard.com/report/github.com/spiral/roadrunner"><img src="https://goreportcard.com/badge/github.com/spiral/roadrunner"></a>
	<a href="https://scrutinizer-ci.com/g/spiral/roadrunner/?branch=master"><img src="https://scrutinizer-ci.com/g/spiral/roadrunner/badges/quality-score.png"></a>
	<a href="https://codecov.io/gh/spiral/roadrunner/"><img src="https://codecov.io/gh/spiral/roadrunner/branch/master/graph/badge.svg"></a>
	<a href="https://lgtm.com/projects/g/spiral/roadrunner/alerts/"><img alt="Total alerts" src="https://img.shields.io/lgtm/alerts/g/spiral/roadrunner.svg?logo=lgtm&logoWidth=18"/></a>
	<a href="https://discord.gg/TFeEmCs"><img src="https://img.shields.io/badge/discord-chat-magenta.svg"></a>
</p>

RoadRunner is an open-source (MIT licensed) high-performance PHP application server, load balancer, and process manager.
It supports running as a service with the ability to extend its functionality on a per-project basis. 

RoadRunner includes PSR-7/PSR-17 compatible HTTP and HTTP/2 server and can be used to replace classic Nginx+FPM setup with much greater performance and flexibility.

<p align="center">
	<a href="https://roadrunner.dev/"><b>Official Website</b></a> | 
	<a href="https://roadrunner.dev/docs"><b>Documentation</b></a>
</p>

Features:
--------
- Production-ready
- PCI DSS compliant
- PSR-7 HTTP server (file uploads, error handling, static files, hot reload, middlewares, event listeners)
- HTTPS and HTTP/2 support (including HTTP/2 Push, H2C)
- Fully customizable server, FastCGI support
- Flexible environment configuration
- No external PHP dependencies (64bit version required), drop-in (based on [Goridge](https://github.com/spiral/goridge))
- Load balancer, process manager and task pipeline
- Frontend agnostic ([Queue](https://github.com/spiral/jobs), PSR-7, [GRPC](https://github.com/spiral/php-grpc), etc)
- Integrated metrics (Prometheus)
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
- Very fast (~250k rpc calls per second on Ryzen 1700X using 16 threads)
- Integrations with Symfony, [Laravel](https://github.com/spiral/roadrunner-laravel), Slim, CakePHP, Zend Expressive
- Application server for [Spiral](https://github.com/spiral/framework)
- Automatic reloading on file changes
- Works on Windows (Unix sockets (AF_UNIX) supported on Windows 10)

Installation:
--------
To install:

```
$ composer require spiral/roadrunner
$ ./vendor/bin/rr get-binary
```

> For getting roadrunner binary file you can use our docker image: `spiralscout/roadrunner:X.X.X` (more information about image and tags can be found [here](https://hub.docker.com/r/spiralscout/roadrunner/))

Extensions:
--------
| Extension | Current Status        
| ---       | ---
spiral/jobs | [![Latest Stable Version](https://poser.pugx.org/spiral/jobs/version)](https://packagist.org/packages/spiral/jobs) [![Build Status](https://travis-ci.org/spiral/jobs.svg?branch=master)](https://travis-ci.org/spiral/jobs) [![Codecov](https://codecov.io/gh/spiral/jobs/branch/master/graph/badge.svg)](https://codecov.io/gh/spiral/jobs/)
spiral/php-grpc | [![Latest Stable Version](https://poser.pugx.org/spiral/php-grpc/version)](https://packagist.org/packages/spiral/php-grpc) [![Build Status](https://travis-ci.org/spiral/php-grpc.svg?branch=master)](https://travis-ci.org/spiral/php-grpc) [![Codecov](https://codecov.io/gh/spiral/php-grpc/branch/master/graph/badge.svg)](https://codecov.io/gh/spiral/php-grpc/)
spiral/broadcast | [![Latest Stable Version](https://poser.pugx.org/spiral/broadcast/version)](https://packagist.org/packages/spiral/broadcast) [![Build Status](https://travis-ci.org/spiral/broadcast.svg?branch=master)](https://travis-ci.org/spiral/broadcast) [![Codecov](https://codecov.io/gh/spiral/broadcast/branch/master/graph/badge.svg)](https://codecov.io/gh/spiral/broadcast/)
spiral/broadcast-ws | [![Latest Stable Version](https://poser.pugx.org/spiral/broadcast-ws/version)](https://packagist.org/packages/spiral/broadcast-ws) [![Build Status](https://travis-ci.org/spiral/broadcast-ws.svg?branch=master)](https://travis-ci.org/spiral/broadcast-ws) [![Codecov](https://codecov.io/gh/spiral/broadcast-ws/branch/master/graph/badge.svg)](https://codecov.io/gh/spiral/broadcast-ws/)

Example:
--------

```php
<?php
// worker.php
ini_set('display_errors', 'stderr');
include "vendor/autoload.php";

$relay = new Spiral\Goridge\StreamRelay(STDIN, STDOUT);
$psr7 = new Spiral\RoadRunner\PSR7Client(new Spiral\RoadRunner\Worker($relay));

while ($req = $psr7->acceptRequest()) {
    try {
        $resp = new \Zend\Diactoros\Response();
        $resp->getBody()->write("hello world");

        $psr7->respond($resp);
    } catch (\Throwable $e) {
        $psr7->getWorker()->error((string)$e);
    }
}
```

Configuration can be located in `.rr.yaml` file ([full sample](https://github.com/spiral/roadrunner/blob/master/.rr.yaml)):

```yaml
http:
  address:         0.0.0.0:8080
  workers.command: "php worker.php"
```

> Read more in [Documentation](https://roadrunner.dev/docs).

Run:
----
To run application server:

```
$ ./rr serve -v -d
```

License:
--------
The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information. Maintained by [Spiral Scout](https://spiralscout.com).
