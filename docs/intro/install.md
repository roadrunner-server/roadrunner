# Installation
The easiest way to get the latest RoadRunner version is to use one of the pre-built release binaries which are available for
OSX, Linux, FreeBSD, and Windows. Instructions for using these binaries are on the GitHub [releases page](https://github.com/spiral/roadrunner-binary/releases).

#### Installation via Composer
You can also install RoadRunner automatically using command shipped with the composer package, run:

```bash
$ composer require spiral/roadrunner
$ ./vendor/bin/rr get
```

Server binary will be available in the root of your project.

> PHP's extensions `php-curl` and `php-zip` are required to download RoadRunner automatically.
> PHP's extensions `php-sockets` need to be installed to run roadrunner.
> Check with `php --modules` your installed extensions.

#### Building RoadRunner
RoadRunner can be compiled on Linux, OSX, Windows and other 64 bit environments as the only requirement are **Go 1.13+**.

To get all needed dependencies:

```bash
$ go mod download
```

To build:

```bash
$ make
```

To test:

```
$ make test
```
