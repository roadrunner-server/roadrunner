RoadRunner
==========
[![Latest Stable Version](https://poser.pugx.org/spiral/roadrunner/version)](https://packagist.org/packages/spiral/roadrunner)
[![GoDoc](https://godoc.org/github.com/spiral/roadrunner?status.svg)](https://godoc.org/github.com/spiral/roadrunner)
[![Build Status](https://travis-ci.org/spiral/roadrunner.svg?branch=master)](https://travis-ci.org/spiral/roadrunner)
[![Go Report Card](https://goreportcard.com/badge/github.com/spiral/roadrunner)](https://goreportcard.com/report/github.com/spiral/roadrunner)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/spiral/roadrunner/badges/quality-score.png)](https://scrutinizer-ci.com/g/spiral/roadrunner/?branch=master)
[![Codecov](https://codecov.io/gh/spiral/roadrunner/branch/master/graph/badge.svg)](https://codecov.io/gh/spiral/roadrunner/)

RoadRunner is an open source (MIT licensed), high-performance PHP application server, load balancer and process manager.
It supports running as a service with the ability to extend its functionality on a per-project basis. 

RoadRunner includes PSR-7 compatible HTTP server and can be used to replace classic Nginx+FPM setup with much greater performance.

Table of Contents 
-----------------
* Introduction
  * [About RoadRunner](https://github.com/spiral/roadrunner/wiki/About-RoadRunner)
  * [Installation](https://github.com/spiral/roadrunner/wiki/Installation)
  * [Configuration](https://github.com/spiral/roadrunner/wiki/Configuration)
  * [License](https://github.com/spiral/roadrunner/wiki/License)
* Using RoadRunner
  * [Environment Configuration](https://github.com/spiral/roadrunner/wiki/Enviroment-Configuration)
  * [HTTPS and HTTP/2](https://github.com/spiral/roadrunner/wiki/HTTPS-and-HTTP2)
  * [PHP Workers](https://github.com/spiral/roadrunner/wiki/PHP-Workers)
  * [Server Commands](https://github.com/spiral/roadrunner/wiki/Server-Commands)
  * [RPC Integration](https://github.com/spiral/roadrunner/wiki/RPC-Integration)
  * [IDE integration](https://github.com/spiral/roadrunner/wiki/IDE-Integration)
  * [Error Handling](https://github.com/spiral/roadrunner/wiki/Debug-And-Error-Handling)
  * [Production Usage](https://github.com/spiral/roadrunner/wiki/Production-Usage)
* Server Customization
  * [Building Server](https://github.com/spiral/roadrunner/wiki/Building-Server)
  * [Writing Services](https://github.com/spiral/roadrunner/wiki/Writing-Services)
  * [HTTP Middlewares](https://github.com/spiral/roadrunner/wiki/Middlewares)
* Additional Notes
  * [Event Listeners](https://github.com/spiral/roadrunner/wiki/Event-Listeners)
  * [Standalone Usage](https://github.com/spiral/roadrunner/wiki/Standalone-usage)
  * [AWS Lambda](https://github.com/spiral/roadrunner/wiki/AWS-Lambda)

Features:
--------
- production ready
- PSR-7 HTTP server (file uploads, error handling, static files, hot reload, middlewares, event listeners)
- HTTPS and HTTP/2 support (including HTTP/2 Push)
- fully customizable server
- flexible environment configuration
- no external PHP dependencies, drop-in (based on [Goridge](https://github.com/spiral/goridge))
- load balancer, process manager and task pipeline
- frontend agnostic (queue, REST, PSR-7, async php, etc)
- works over TCP, unix sockets and standard pipes
- automatic worker replacement and safe PHP process destruction
- worker lifecycle management (create/allocate/destroy timeouts)
- payload context and body
- control over max jobs per worker
- protocol, worker and job level error management (including PHP errors)
- memory leak failswitch
- very fast (~250k rpc calls per second on Ryzen 1700X using 16 threads)
- works on Windows

Example:
--------

```php
<?php
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

Configuration can be located in `.rr.yaml` file:

```yaml
http:
  address:    0.0.0.0:8080
  maxRequest: 200
  workers:
    command: "php psr-worker.php"
    pool:
      numWorkers: 4
```

License:
--------
The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information. Maintained by [SpiralScout](https://spiralscout.com).
