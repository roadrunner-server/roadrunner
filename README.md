<p align="center">
 <img src="https://user-images.githubusercontent.com/796136/50286124-6f7f3780-046f-11e9-9f45-e8fedd4f786d.png" height="75px" alt="RoadRunner">
</p>
<p align="center">
 <a href="https://packagist.org/packages/spiral/roadrunner"><img src="https://poser.pugx.org/spiral/roadrunner/version"></a>
	<a href="https://godoc.org/github.com/spiral/roadrunner"><img src="https://godoc.org/github.com/spiral/roadrunner?status.svg"></a>
	<a href="https://travis-ci.org/spiral/roadrunner"><img src="https://travis-ci.org/spiral/roadrunner.svg?branch=master"></a>
	<a href="https://goreportcard.com/report/github.com/spiral/roadrunner"><img src="https://goreportcard.com/badge/github.com/spiral/roadrunner"></a>
	<a href="https://scrutinizer-ci.com/g/spiral/roadrunner/?branch=master"><img src="https://scrutinizer-ci.com/g/spiral/roadrunner/badges/quality-score.png"></a>
	<a href="https://codecov.io/gh/spiral/roadrunner/"><img src="https://codecov.io/gh/spiral/roadrunner/branch/master/graph/badge.svg"></a>
	<a href="https://discord.gg/TFeEmCs"><img src="https://img.shields.io/badge/discord-chat-magenta.svg"></a>
</p>

RoadRunner is an open source (MIT licensed) high-performance PHP application server, load balancer and process manager.
It supports running as a service with the ability to extend its functionality on a per-project basis. 

RoadRunner includes PSR-7/PSR-17 compatible HTTP and HTTP/2 server and can be used to replace classic Nginx+FPM setup with much greater performance and flexibility.

<p align="center">
	<a href="https://roadrunner.dev/"><b>Official Website</b></a> | 
	<a href="https://roadrunner.dev/docs"><b>Documentation</b></a>
</p>

Features:
--------
- production ready
- PSR-7 HTTP server (file uploads, error handling, static files, hot reload, middlewares, event listeners)
- HTTPS and HTTP/2 support (including HTTP/2 Push)
- fully customizable server
- flexible environment configuration
- no external PHP dependencies (64bit version required), drop-in (based on [Goridge](https://github.com/spiral/goridge))
- load balancer, process manager and task pipeline
- frontend agnostic ([Queue](https://github.com/spiral/jobs), PSR-7, [GRPC](https://github.com/spiral/php-grpc), etc)
- integrated metrics (Prometheus)
- works over TCP, unix sockets and standard pipes
- automatic worker replacement and safe PHP process destruction
- worker create/allocate/destroy timeouts
- max jobs per worker
- worker lifecycle management (controller) 
    - maxMemory (graceful stop)
    - TTL (graceful stop)
    - idleTTL (graceful stop)
    - execTTL (brute, max_execution_time)   
- payload context and body
- protocol, worker and job level error management (including PHP errors)
- very fast (~250k rpc calls per second on Ryzen 1700X using 16 threads)
- integrations with Symfony, Laravel, Slim, CakePHP, Zend Expressive, Spiral
- works on Windows

Installation:
--------
To install:

```
$ composer require spiral/roadrunner
$ ./vendor/bin/rr get-binary
```

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

> Read more in [Wiki](https://github.com/spiral/roadrunner/wiki/PHP-Workers).

Run:
----
To run application server:

```
$ ./rr serve -v -d
```

License:
--------
The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information. Maintained by [Spiral Scout](https://spiralscout.com).
