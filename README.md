<p align="center">
 <img src="https://user-images.githubusercontent.com/796136/50286124-6f7f3780-046f-11e9-9f45-e8fedd4f786d.png" height="75px" alt="RoadRunner">
</p>
<p align="center">
 <a href="https://packagist.org/packages/spiral/roadrunner"><img src="https://poser.pugx.org/spiral/roadrunner/version"></a>
	<a href="https://pkg.go.dev/github.com/roadrunner-server/roadrunner/v2?tab=doc"><img src="https://godoc.org/github.com/roadrunner-server/roadrunner/v2?status.svg"></a>
    <a href="https://codecov.io/gh/roadrunner-server/roadrunner/"><img src="https://codecov.io/gh/roadrunner-server/roadrunner/branch/master/graph/badge.svg"></a>
	<a href="https://github.com/roadrunner-server/roadrunner/actions"><img src="https://github.com/roadrunner-server/roadrunner/workflows/rr_cli_tests/badge.svg" alt=""></a>
    <a href="https://codecov.io/gh/roadrunner-server/rr-e2e-tests/"><img src="https://codecov.io/gh/roadrunner-server/rr-e2e-tests/branch/master/graph/badge.svg"></a>
    <a href="https://github.com/roadrunner-server/rr-e2e-tests/actions"><img src="https://github.com/roadrunner-server/rr-e2e-tests/workflows/linux_stable/badge.svg" alt=""></a>
	<a href="https://github.com/roadrunner-server/rr-e2e-tests/actions"><img src="https://github.com/roadrunner-server/rr-e2e-tests/workflows/Linters/badge.svg" alt=""></a>
	<a href="https://goreportcard.com/report/github.com/roadrunner-server/roadrunner"><img src="https://goreportcard.com/badge/github.com/roadrunner-server/roadrunner"></a>
	<a href="https://lgtm.com/projects/g/roadrunner-server/roadrunner/alerts/"><img alt="Total alerts" src="https://img.shields.io/lgtm/alerts/g/roadrunner-server/roadrunner.svg?logo=lgtm&logoWidth=18"/></a>
	<a href="https://discord.gg/TFeEmCs"><img src="https://img.shields.io/badge/discord-chat-magenta.svg"></a>
	<a href="https://packagist.org/packages/spiral/roadrunner"><img src="https://img.shields.io/packagist/dd/spiral/roadrunner?style=flat-square"></a>
    <img alt="All releases" src="https://img.shields.io/github/downloads/roadrunner-server/roadrunner/total">
</p>

RoadRunner is an open-source (MIT licensed) high-performance PHP application server, load balancer, and process manager.
It supports running as a service with the ability to extend its functionality on a per-project basis.

RoadRunner includes PSR-7/PSR-17 compatible HTTP and HTTP/2 server and can be used to replace classic Nginx+FPM setup
with much greater performance and flexibility.

## Join our discord server: [Link](https://discord.gg/TFeEmCs)

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
- No external PHP dependencies (64bit version required), drop-in (based on [Goridge](https://github.com/roadrunner-server/goridge))
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

To get the roadrunner binary file you can use our docker image: `spiralscout/roadrunner:2.X.X` (more information about
image and tags can be found [here](https://hub.docker.com/r/spiralscout/roadrunner/)) or use the GitHub package: `ghcr.io/roadrunner-server/roadrunner:2.X.X`


- Docker:

```dockerfile
FROM ghcr.io/roadrunner-server/roadrunner:2.X.X AS roadrunner
FROM php:8.1-cli

COPY --from=roadrunner /usr/bin/rr /usr/local/bin/rr

# USE THE RR
```

- CLI

```bash
$ composer require spiral/roadrunner:v2.0 nyholm/psr7
$ ./vendor/bin/rr get-binary
```


Configuration can be located in `.rr.yaml`
file ([full sample](https://github.com/roadrunner-server/roadrunner/blob/master/.rr.yaml)):

```yaml
# configuration version: https://roadrunner.dev/docs/beep-beep-config/2.x/en
version: '2.7'

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

# Available Plugins:

| Plugin                                             | Description                                                                                                                         | Latest tag                                                            | Go version                                                                        |
|----------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------|-----------------------------------------------------------------------------------|
| ![](https://img.shields.io/badge/-HTTP-grey)       | Provides HTTP, HTTPS, FCGI transports. Extensible with middleware.                                                                  | ![](https://img.shields.io/github/v/tag/roadrunner-server/http)       | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/http)       |
| ![](https://img.shields.io/badge/-Headers-grey)    | HTTP middleware supports constant custom headers and CORS.                                                                          | ![](https://img.shields.io/github/v/tag/roadrunner-server/headers)    | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/headers)    |
| ![](https://img.shields.io/badge/-GZIP-grey)       | HTTP middleware supports `Accept-Encoding`: GZIP.                                                                                   | ![](https://img.shields.io/github/v/tag/roadrunner-server/gzip)       | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/gzip)       |
| ![](https://img.shields.io/badge/-Static-grey)     | HTTP middleware serves static files.                                                                                                | ![](https://img.shields.io/github/v/tag/roadrunner-server/static)     | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/static)     |
| ![](https://img.shields.io/badge/-Sendfile-grey)   | HTTP middleware handles X-Sendfile headers.                                                                                         | ![](https://img.shields.io/github/v/tag/roadrunner-server/send)       | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/send)       |
| ![](https://img.shields.io/badge/-NewRelic-grey)   | HTTP middleware supports NewRelic distributed traces and custom attributes.                                                         | ![](https://img.shields.io/github/v/tag/roadrunner-server/new_relic)  | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/new_relic)  |
| ![](https://img.shields.io/badge/-Cache-grey)      | HTTP middleware supports RFC7234 cache.                                                                                             | ![](https://img.shields.io/github/v/tag/roadrunner-server/cache)      | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/cache)      |
| ![](https://img.shields.io/badge/-Cache-grey)      | HTTP middleware supports OpenTelemetry protocol.                                                                                    | ![](https://img.shields.io/github/v/tag/roadrunner-server/otel)      | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/otel)      |
| ![](https://img.shields.io/badge/-Jobs-grey)       | Provides queues support for the RR2 via different drivers                                                                           | ![](https://img.shields.io/github/v/tag/roadrunner-server/jobs)       | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/jobs)       |
| ![](https://img.shields.io/badge/-AMQP-grey)       | Provides AMQP (0-9-1) protocol support via RabbitMQ                                                                                 | ![](https://img.shields.io/github/v/tag/roadrunner-server/amqp)       | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/amqp)       |
| ![](https://img.shields.io/badge/-Beanstalk-grey)  | Provides [beanstalkd](https://github.com/beanstalkd/beanstalkd) queue support                                                       | ![](https://img.shields.io/github/v/tag/roadrunner-server/beanstalk)  | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/beanstalk)  |
| ![](https://img.shields.io/badge/-Boltdb-grey)     | Provides support for the [BoltDB](https://github.com/etcd-io/bbolt) key/value store. Used in the `Jobs` and `KV`                    | ![](https://img.shields.io/github/v/tag/roadrunner-server/boltdb)     | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/boltdb)     |
| ![](https://img.shields.io/badge/-SQS-grey)        | SQS driver for the jobs                                                                                                             | ![](https://img.shields.io/github/v/tag/roadrunner-server/sqs)        | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/sqs)        |
| ![](https://img.shields.io/badge/-NATS-grey)       | NATS jobs driver                                                                                                                    | ![](https://img.shields.io/github/v/tag/roadrunner-server/nats)       | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/nats)       |
| ![](https://img.shields.io/badge/-KV-grey)         | Provides key-value support for the RR2 via different drivers                                                                        | ![](https://img.shields.io/github/v/tag/roadrunner-server/kv)         | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/kv)         |
| ![](https://img.shields.io/badge/-Memcached-grey)  | Memcached driver for the kv                                                                                                         | ![](https://img.shields.io/github/v/tag/roadrunner-server/memcached)  | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/memcached)  |
| ![](https://img.shields.io/badge/-Memory-grey)     | Memory driver for the jobs, kv, broadcast                                                                                           | ![](https://img.shields.io/github/v/tag/roadrunner-server/memory)     | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/memory)     |
| ![](https://img.shields.io/badge/-Redis-grey)      | Redis driver for the kv, broadcast                                                                                                  | ![](https://img.shields.io/github/v/tag/roadrunner-server/redis)      | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/redis)      |
| ![](https://img.shields.io/badge/-Config-grey)     | Provides configuration parsing support to the all plugins                                                                           | ![](https://img.shields.io/github/v/tag/roadrunner-server/config)     | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/config)     |
| ![](https://img.shields.io/badge/-GRPC-grey)       | Provides GRPC support                                                                                                               | ![](https://img.shields.io/github/v/tag/roadrunner-server/grpc)       | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/grpc)       |
| ![](https://img.shields.io/badge/-Informer-grey)   | Provides statistic grabbing capabilities (workers,jobs stat)                                                                        | ![](https://img.shields.io/github/v/tag/roadrunner-server/informer)   | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/informer)   |
| ![](https://img.shields.io/badge/-Broadcast-grey)  | Provides broadcasting capabilities to the RR2 via different drivers                                                                 | ![](https://img.shields.io/github/v/tag/roadrunner-server/broadcast)  | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/broadcast)  |
| ![](https://img.shields.io/badge/-Logger-grey)     | Central logger plugin. Implemented via Uber.zap logger, but supports other loggers.                                                 | ![](https://img.shields.io/github/v/tag/roadrunner-server/logger)     | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/logger)     |
| ![](https://img.shields.io/badge/-Metrics-grey)    | Provides support for the metrics via [Prometheus](https://prometheus.io/)                                                           | ![](https://img.shields.io/github/v/tag/roadrunner-server/metrics)    | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/metrics)    |
| ![](https://img.shields.io/badge/-Reload-grey)     | Reloads workers on the file changes. Use only for the development                                                                   | ![](https://img.shields.io/github/v/tag/roadrunner-server/reload)     | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/reload)     |
| ![](https://img.shields.io/badge/-Resetter-grey)   | Provides support for the `./rr reset` command. Reloads workers pools                                                                | ![](https://img.shields.io/github/v/tag/roadrunner-server/resetter)   | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/resetter)   |
| ![](https://img.shields.io/badge/-RPC-grey)        | Provides support for the RPC across all plugins. Collects `RPC() interface{}` methods and exposes them via RPC                      | ![](https://img.shields.io/github/v/tag/roadrunner-server/rpc)        | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/rpc)        |
| ![](https://img.shields.io/badge/-Server-grey)     | Provides support for the command. Prepare PHP processes                                                                             | ![](https://img.shields.io/github/v/tag/roadrunner-server/server)     | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/server)     |
| ![](https://img.shields.io/badge/-Service-grey)    | Provides support for the external scripts, binaries which might be started like a service (behaves similar to the systemd services) | ![](https://img.shields.io/github/v/tag/roadrunner-server/service)    | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/service)    |
| ![](https://img.shields.io/badge/-Status-grey)     | Provides support for the health and readiness checks                                                                                | ![](https://img.shields.io/github/v/tag/roadrunner-server/status)     | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/status)     |
| ![](https://img.shields.io/badge/-Websockets-grey) | Provides support for the broadcasting events via websockets                                                                         | ![](https://img.shields.io/github/v/tag/roadrunner-server/websockets) | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/websockets) |
| ![](https://img.shields.io/badge/-TCP-grey)        | Provides support for the raw TCP payloads and TCP connections                                                                       | ![](https://img.shields.io/github/v/tag/roadrunner-server/tcp)        | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/tcp)        |
| ![](https://img.shields.io/badge/-Fileserver-grey) | File server to handle static files                                                                                                  | ![](https://img.shields.io/github/v/tag/roadrunner-server/fileserver) | ![](https://img.shields.io/github/go-mod/go-version/roadrunner-server/fileserver) |

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

## Contributors

Thanks to all the people who already contributed!

<a href="https://github.com/roadrunner-server/roadrunner/graphs/contributors">
  <img src="https://contributors-img.web.app/image?repo=roadrunner-server/roadrunner" />
</a>
