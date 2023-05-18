<p align="center">
 <a href="https://roadrunner.dev" target="_blank">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://user-images.githubusercontent.com/7326800/205905278-3899e2c8-5c15-4347-820b-a8ea4c5ba2d7.png">
    <img align="center" src="https://user-images.githubusercontent.com/796136/50286124-6f7f3780-046f-11e9-9f45-e8fedd4f786d.png">
  </picture>
</a>
</p>
<p align="center">
 <a href="https://packagist.org/packages/spiral/roadrunner"><img src="https://poser.pugx.org/spiral/roadrunner/version"></a>
	<a href="https://pkg.go.dev/github.com/roadrunner-server/roadrunner/v2023?tab=doc"><img src="https://godoc.org/github.com/roadrunner-server/roadrunner/v2023?status.svg"></a>
    <a href="https://twitter.com/spiralphp"><img src="https://img.shields.io/twitter/follow/spiralphp?style=social"></a>
    <a href="https://codecov.io/gh/roadrunner-server/roadrunner/"><img src="https://codecov.io/gh/roadrunner-server/roadrunner/branch/master/graph/badge.svg"></a>
	<a href="https://github.com/roadrunner-server/roadrunner/actions"><img src="https://github.com/roadrunner-server/roadrunner/workflows/rr_cli_tests/badge.svg" alt=""></a>
    <a href="https://codecov.io/gh/roadrunner-server/rr-e2e-tests/"><img src="https://codecov.io/gh/roadrunner-server/rr-e2e-tests/branch/master/graph/badge.svg"></a>
    <a href="https://github.com/roadrunner-server/rr-e2e-tests/actions"><img src="https://github.com/roadrunner-server/rr-e2e-tests/workflows/linux_on_push/badge.svg" alt=""></a>
	<a href="https://github.com/roadrunner-server/rr-e2e-tests/actions"><img src="https://github.com/roadrunner-server/rr-e2e-tests/workflows/Linters/badge.svg" alt=""></a>
	<a href="https://goreportcard.com/report/github.com/roadrunner-server/roadrunner/v2"><img src="https://goreportcard.com/badge/github.com/roadrunner-server/roadrunner/v2"></a>
	<a href="https://discord.gg/TFeEmCs"><img src="https://img.shields.io/badge/discord-chat-magenta.svg"></a>
	<a href="https://packagist.org/packages/spiral/roadrunner"><img src="https://img.shields.io/packagist/dd/spiral/roadrunner?style=flat-square"></a>
    <img alt="All releases" src="https://img.shields.io/github/downloads/roadrunner-server/roadrunner/total">
</p>

RoadRunner is an open-source (MIT licensed) high-performance PHP application server, load balancer, and process manager.
It supports running as a service with the ability to extend its functionality on a per-project basis.

RoadRunner includes PSR-7/PSR-17 compatible HTTP and HTTP/2 server and can be used to replace classic Nginx+FPM setup
with much greater performance and flexibility.

# Join our discord server: [Link](https://discord.gg/TFeEmCs)

<p align="center">
	<a href="https://roadrunner.dev/"><b>Official Website</b></a> |
	<a href="https://roadrunner.dev/docs"><b>Documentation</b></a> |
    <a href="https://forum.roadrunner.dev"><b>Forum</b></a> |
    <a href="https://github.com/orgs/roadrunner-server/projects/1"><b>Release schedule</b></a>
</p>

# Features:
--------
- Production-ready
- PCI DSS compliant (HTTP plugin)
- PSR-7 HTTP server (file uploads, error handling, static files, hot reload, middleware, event listeners)
- HTTPS and HTTP/2 support (including HTTP/2 Push, H2C)
- A Fully customizable http(s)/2 server
- FastCGI support (HTTP plugin)
- Flexible environment configuration
- No external PHP dependencies (64bit version required)
- Integrated metrics (Prometheus)
- [Workflow engine](https://github.com/temporalio/sdk-php) by [Temporal.io](https://temporal.io)
- Works over TCP, UNIX sockets and process pipes
- Automatic worker replacement, graceful and safe PHP process destruction
- Worker create/allocate/destroy timeouts
- Max requests per worker limitation
- Worker lifecycle management (controller)
    - max_memory (graceful stop)
    - ttl (graceful stop)
    - idle_ttl (graceful stop)
    - exec_tll (brute, max_execution_time)
- Protocol, worker and job level error management (including PHP errors)
- Development Mode
- Integrations with [Symfony](https://roadrunner.dev/docs/integration-symfony), [Laravel](https://roadrunner.dev/docs/integration-laravel), Slim, CakePHP, Zend Expressive
- Application server for [Spiral](https://github.com/spiral/framework)
- Works on Windows 10/11 and WSL2

---

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
You can also install RoadRunner automatically using command shipped with the composer package, run:

```bash
composer require spiral/roadrunner-http nyholm/psr7
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

## Dowload the latest release via curl:
```bash
curl --proto '=https' --tlsv1.2 -sSf  https://raw.githubusercontent.com/roadrunner-server/roadrunner/master/download-latest.sh | sh
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
