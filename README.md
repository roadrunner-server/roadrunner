<p align="center">
 <a href="https://roadrunner.dev" target="_blank">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/roadrunner-server/.github/assets/8040338/cf1bfcf2-b787-426d-80f5-2862bb2a39b2">
    <img align="center" src="https://github.com/roadrunner-server/.github/assets/8040338/c4b971fd-b84f-406d-b850-0a4f072a5885">
  </picture>
</a>
</p>
<p align="center">
 <a href="https://packagist.org/packages/spiral/roadrunner"><img src="https://poser.pugx.org/spiral/roadrunner/version"></a>
	<a href="https://pkg.go.dev/github.com/roadrunner-server/roadrunner/v2023?tab=doc"><img src="https://godoc.org/github.com/roadrunner-server/roadrunner/v2023?status.svg"></a>
    <a href="https://twitter.com/spiralphp"><img src="https://img.shields.io/twitter/follow/spiralphp?style=social"></a>
    <a href="https://codecov.io/gh/roadrunner-server/roadrunner/"><img src="https://codecov.io/gh/roadrunner-server/roadrunner/branch/master/graph/badge.svg"></a>
	<a href="https://github.com/roadrunner-server/roadrunner/actions"><img src="https://github.com/roadrunner-server/roadrunner/workflows/rr_cli_tests/badge.svg" alt=""></a>
	<a href="https://goreportcard.com/report/github.com/roadrunner-server/roadrunner/v2"><img src="https://goreportcard.com/badge/github.com/roadrunner-server/roadrunner/v2"></a>
	<a href="https://discord.gg/TFeEmCs"><img src="https://img.shields.io/badge/discord-chat-magenta.svg"></a>
	<a href="https://packagist.org/packages/spiral/roadrunner"><img src="https://img.shields.io/packagist/dd/spiral/roadrunner?style=flat-square"></a>
    <img alt="All releases" src="https://img.shields.io/github/downloads/roadrunner-server/roadrunner/total">
</p>

RoadRunner is an open-source (MIT licensed) high-performance PHP application server, process manager written in Go and powered with plugins ❤️.
It supports running as a service with the ability to extend its functionality on a per-project basis with plugins.


# Features

**RoadRunner** features a range of plugins, including `HTTP(S)/2/3` and `fCGI` servers that are compatible with PSR-7/PSR-17 standards. This is just one of its many capabilities. It serves as an effective alternative to the traditional Nginx+FPM setup, providing improved performance and more flexibility. Its extensive plugin options go far beyond just `HTTP(S)/2/3` and `fCGI` servers, offering a broad range of functionalities:
- Queue drivers: RabbitMQ, Kafka, SQS, Beanstalk, NATS, In-Memory.
- KV drivers: Redis, Memcached, BoltDB, In-Memory.
- OpenTelemetry protocol support (`gRPC`, `http`, `jaeger`).
- [Workflow engine](https://github.com/temporalio/sdk-php) via [Temporal](https://temporal.io).
- `gRPC` server. For increased speed, the `protobuf` extension can be used.
- `HTTP(S)/2/3` and `fCGI` servers features **automatic TLS management**, **103 Early Hints** support and middleware like: Static, Headers, gzip, prometheus (metrics), send (x-sendfile), OTEL, proxy_ip_parser, etc.
- Embedded distribute lock plugin which manages access to shared resources.
- Metrics server (you might easily expose your own).
- WebSockets and Broadcast via [Centrifugo](https://centrifugal.dev) server.
- Systemd-like services manager with auto-restarts, execution time limiter, etc.
- Production-ready.
- And more 😉

# Join our discord server: [Link](https://discord.gg/TFeEmCs)

<p align="center">
	<a href="https://roadrunner.dev/"><b>Official Website</b></a> |
	<a href="https://docs.roadrunner.dev/"><b>Documentation</b></a> |
    <a href="https://forum.roadrunner.dev"><b>Forum</b></a> |
    <a href="https://github.com/orgs/roadrunner-server/projects/4"><b>Release schedule</b></a>
</p>

# Installation

The easiest way to get the latest RoadRunner version is to use one of the pre-built release binaries, which are available for
OSX, Linux, FreeBSD, and Windows. Instructions for using these binaries are on the GitHub [releases page](https://github.com/roadrunner-server/roadrunner/releases).

## Docker:

To get the roadrunner binary file you can use our docker image: `ghcr.io/roadrunner-server/roadrunner:2023.X.X` (more information about
image and tags can be found [here](https://github.com/roadrunner-server/roadrunner/pkgs/container/roadrunner)).

```dockerfile
FROM ghcr.io/roadrunner-server/roadrunner:2023.X.X AS roadrunner
FROM php:8.2-cli

COPY --from=roadrunner /usr/bin/rr /usr/local/bin/rr

# USE THE RR
```

Configuration located in the `.rr.yaml` file ([full sample](https://github.com/roadrunner-server/roadrunner/blob/master/.rr.yaml)):


## Installation via Composer
You can also install RoadRunner automatically using the command shipped with the composer package, run:

```bash
composer require spiral/roadrunner-http spiral/roadrunner-worker nyholm/psr7
./vendor/bin/rr get-binary
```

Server binary will be available at the root of your project.

> **Note**
>
> PHP's extensions `php-curl` and `php-zip` are required to download RoadRunner automatically.
> PHP's extensions `php-sockets` need to be installed to run roadrunner.
> Check with `php --modules` your installed extensions.


## Installation option for the Debian-derivatives (Ubuntu, Mint, MX, etc)

```bash
wget https://github.com/roadrunner-server/roadrunner/releases/download/v2023.X.X/roadrunner-2023.X.X-linux-amd64.deb
sudo dpkg -i roadrunner-2023.X.X-linux-amd64.deb
```

## Download the latest release via curl:
```bash
curl --proto '=https' --tlsv1.2 -sSf  https://raw.githubusercontent.com/roadrunner-server/roadrunner/master/download-latest.sh | sh
```

## MacOS using [Homebrew](https://brew.sh/):
```bash
brew install roadrunner
```

---

Configuration can be located in `.rr.yaml` file ([full sample](https://github.com/roadrunner-server/roadrunner/blob/master/.rr.yaml)):

```yaml
version: '3'

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

> [!IMPORTANT]
> If you see the `EOF` error, check that you have installed the PHP packages from [this step](https://github.com/roadrunner-server/roadrunner#installation-via-composer).
> If this does not help, try to execute the command `php worker.php` directly and check the output.

---

### Available Plugins: [link](https://roadrunner.dev/docs)

Run:
----
To run application server:

```
$ ./rr serve -c .rr.yaml
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
