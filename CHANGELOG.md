CHANGELOG
=========

UNRELEASED
----------
- Add `rr --version` flag support

v1.9.0 (02.12.2020)
-------------------
- Update PHP minimal supported version to 7.3
- Support PHP 8.0
- Update Ubuntu version in GHA to 20.04
- Update Golang version in the RR Dockerfile to 1.15.5

v1.8.4 (21.10.2020)
-------------------
- Update Goridge go dep to 2.4.6

v1.8.3 (02.09.2020)
-------------------
- Fix superfluous response.WriteHeader issue.
- Proper handle of `broken pipe` error on Linux and Windows.
- PCI DSS compliant upgrade (thanks @antonydevanchi).
- Fix HSTS header (thanks @antonydevanchi).
- Add Request and Response headers to static files (thanks @siad007).
- Add user_namespaces check when running RR worker from a particular user.

v1.8.2 (06.06.2020)
-------------------
- Bugfix release

v1.8.1 (23.05.2020)
-------------------
- Update goridge version to 2.4.4
- Fix code warnings from phpstan
- Improve RPC
- Create templates for the Bug reporting and Feature requests
- Move docker images from golang-alpine to regular golang images
- Add support for the CloudFlare CF-Connecting-IP and True-Client-IP headers (thanks @vsychov)
- Add support for the Root CA via the `rootCa` .rr.yaml option
- See the full milestone here: [link](https://github.com/spiral/roadrunner/milestone/11?closed=1)

v1.8.0 (05.05.2020)
-------------------
- Update goridge version to 2.4.0
- Update PHP version to the 7.2 (currently minimum supported)
- See the full milestone here: [link](https://github.com/spiral/roadrunner/milestone/10?closed=1)

v1.7.1 (22.04.2020)
-------------------
- Syscall usage optimized. Now the data is packing and sending via 1 (or 2 in some cases) send_socket calls, instead of 2-3 (by @vvval)
- Unix sockets in Windows (AF_UNIX) now supported.
- Systemd unit file now in the root of the repository. Feel free to read the [docs](https://roadrunner.dev/docs/beep-beep-systemd) about running RR as daemon on Linux based systems.
- Added ability to run the worker process from the particular user on Linux-based systems. Make sure, that the user have the permissions to run the script. See the [config](https://roadrunner.dev/docs/intro-config), option `user`.
- Fixed: vendor directory conflict with golang part of the application. Now php uses vendor_php directory for the dependencies.
- Goridge updated to version 2.3.2.
- Deprecated Zend dependency replaced with Laminas-diactoros.
- See the full log: [Milestone](https://github.com/spiral/roadrunner/milestone/9?closed=1)

v1.7.0 (23.03.2020)
-------------------
- Replaced std encoding/json package with the https://github.com/json-iterator/go

v1.6.4 (14.03.2020)
-------------------
- Fixed bug with RR getting unreasonable without config
- Fixed bug with local paths in panic messages
- Fixed NPE bug with empty `http` config and enabled `gzip` plugin

v1.6.3 (10.03.2020)
-------------------
- Fixed bug with UB when the plugin is failing during start
- Better signals handling
- Rotate ports in tests
- Added BORS to repository [https://bors.tech/], so, now you can use commands from it, like `bors d=@some_user`, `bors try`, or `bors r+`
- Reverted change with `musl-gcc`. We reverted `CGO_ENABLED=0`, so, CGO turned off for all targets and `netgo`, `osuser` etc.. system-dependent packages are not statically linked. Also separate `musl` binary provided.
- macOS temporarily removed from CI
- Added curl dependency to download rr (@dkarlovi)

v1.6.2 (23.02.2020)
-------------------
- added reload module to handle file changes

v1.6.1 (17.02.2020)
-------------------
- When you run ./rr server -d (debug mode), also pprof server will be launched on :6061 port (this is default golang port for pprof) with the default endpoints (see: https://golang.org/pkg/net/http/pprof/)
- Added LDFLAGS "-s" to build.sh --> https://spiralscout.com/blog/golang-software-testing-tips

v1.6.0 (11.02.2020)
-------------------
- Moved to GitHub Actions, thanks to @tarampampam
- New GZIP handler, thanks to @wppd
- Tests stabilization and fix REQUEST_URI for requests through FastCGI, thanks to @marliotto
- Golang modules update and new RPC method to register metrics from the application, thanks to @48d90782
- Deadlock on timer update in error buffer [bugfix], thanks to @camohob

v1.5.3 (23.12.2019)
-------------------
- metric and RPC ports are rotated in tests to avoid false positive
- massive test and source cleanup (more error handlers) by @ValeryPiashchynski
- "Server closed" error has been suppressed
- added the ability to specify any config value via JSON flag `-j`
- minor improvements in Travis pipeline
- bump the minimum TLS version to TLS 1.2
- added `Strict-Transport-Security` header for TLS requests

v1.5.2 (05.12.2019)
-------------------
- added support for symfony/console 5.0 by @coxa
- added support for HTTP2 trailers by @filakhtov

v1.5.1 (22.10.2019)
-------------------
- bugfix: do not halt stop sequence in case of service error

v1.5.0 (12.10.2019)
-------------------
- initial code style fixes by @ScullWM
- added health service for better integration with Kubernetes by @awprice
- added support for payloads in GET methods by @moeinpaki
- dropped support of PHP 7.0 version (you can still use new server binary)

v1.4.8 (06.09.2019)
-------------------
- bugfix in proxy IP resolution by @spudro228 
- `rr get` can now skip binary download if version did not change by
  @drefixs
- bugfix in `rr init-config` and with linux binary download by
  @Hunternnm
- `$_SERVER['REQUEST_URI']` is now being set

v1.4.7 (29.07.2019)
-------------------
- added support for H2C over TCP by @Alex-Bond

v1.4.6 (01.07.2019)
-------------------
- Worker is not final (to allow mocking)
- MatricsInterface added

v1.4.5 (27.06.2019)
-------------------
- added metrics server with Prometheus backend
- ability to push metrics from the application
- expose http service metrics
- expose limit service metrics
- expose generic golang metrics
- HttpClient and Worker marked final

v1.4.4 (25.06.2019)
-------------------
- added "headers" service with the ability to specify request, response and CORS headers by @ovr
- added FastCGI support for HTTP service by @ovr
- added ability to include multiple config files using `include` directive in the configuration

v1.4.3 (03.06.2019)
-------------------
- fixed dependency with Zend Diactoros by @dkuhnert 
- minor refactoring of error reporting by @lda

v1.4.2 (22.05.2019)
-------------------
- bugfix: incorrect RPC method for stop command
- bugfix: incorrect archive extension in /vendor/bin/rr get on linux machines

v1.4.1 (15.05.2019)
-------------------
- constrain service renamed to "limit" to equalize the definition with sample config

v1.4.0 (05.05.2019)
-------------------
- launch of official website https://roadrunner.dev/
- ENV variables in configs (automatic RR_ mapping and manual definition using "${ENV_NAME}" value)
- the ability to safely remove the worker from the pool in runtime
- minor performance improvements
- `real ip` resolution using X-Real-Ip and X-Forwarded-For (+cidr verification) 
- automatic worker lifecycle manager (controller, see [sample config](https://github.com/spiral/roadrunner/blob/master/.rr.yaml))
   - maxMemory (graceful stop)
   - ttl (graceful stop)
   - idleTTL (graceful stop)
   - execTTL (brute, max_execution_time)   
- the ability to stop rr using `rr stop`
- `maxRequest` option has been deprecated in favor of `maxRequestSize`
- `/vendor/bin/rr get` to download rr server binary (symfony/console) by @Alex-Bond
- `/vendor/bin/rr init` to init rr config by @Alex-Bond
- quick builds are no longer supported
- PSR-12
- strict_types=1 added to all php files

v1.3.7 (21.03.2019)
-------------------
- bugfix: Request field ordering with same names #136 

v1.3.6 (21.03.2019)
-------------------
- bugfix: pool did not wait for slow workers to complete while running concurrent load with http:reset command being invoked

v1.3.5 (14.02.2019)
-------------------
- new console flag `l` to define log formatting
    * **color|default** - colorized output
    * **plain**         - disable all colorization
    * **json**          - output as json
- new console flag `w` to specify work dir
- added ability to work without config file when at least one `overwrite` option has been specified
- pool config now sets `numWorkers` equal to number of cores by default (this section can be omitted now)

v1.3.4 (02.02.2019)
-------------------
- bugfix: invalid content type detection for urlencoded form requests with custom encoding by @Alex-Bond

v1.3.3 (31.01.2019)
-------------------
- added HttpClient for faster integrations with non PSR-7 frameworks by @Alex-Bond

v1.3.2 (11.01.2019)
-------------------
- `_SERVER` now exposes headers with HTTP_ prefix (fixing Lravel integration) by @Alex-Bond
- fixed bug causing body payload not being received for custom HTTP methods by @Alex-Bond 

v1.3.1 (11.01.2019)
-------------------
- fixed bug causing static_pool crash when multiple reset requests received at the same time
- added `always` directive to static service config to always service files of specific extension
- added `vendor/bin/rr-build` command to easier compile custom RoadRunner builds 

v1.3.0 (05.01.2019)
-------------------
- added support for zend/diactros 1.0 and 2.0
- removed `http-interop/http-factory-diactoros`
- added `strict_types=1`
- added elapsed time into debug log
- ability to redefine config via flags (example: `rr serve -v -d -o http.workers.pool.numWorkers=1`)
- fixed bug causing child processes die before parent rr (annoying error on windows "worker exit status ....")
- improved stop sequence and graceful exit
- `env.Environment` has been spitted into `env.Setter` and `env.Getter`
- added `env.Copy` method
- config management has been moved out from root command into `utils`
- spf13/viper dependency has been bumped up to 1.3.1
- more tests
- new travis configuration

v1.2.8 (26.12.2018)
-------------------
- bugfix #76 error_log redirect has been disabled after `http:reset` command

v1.2.7 (20.12.2018)
-------------------
- #67 bugfix, invalid protocol version while using HTTP/2 with new http-interop by @bognerf
- #66 added HTTP_USER_AGENT value and tests for it
- typo fix in static service by @Alex-Bond
- added PHP 7.3 to travis
- less ambiguous error when invalid data found in a pipe(`invalid prefix (checksum)` => `invalid data found in the buffer (possible echo)`)

v1.2.6 (18.10.2018)
-------------------
- bugfix: ignored `stopping` value during http server shutdown
- debug log now split message into individual lines

v1.2.5 (13.10.2018)
------
- decoupled from Zend Diactoros via PSR-17 factory (by @1ma)
- `Verbose` flag for cli renamed to `verbose` (by @ruudk)
- bugfix: HTTP protocol version mismatch on PHP end

v1.2.4 (30.09.2018)
------
- minor performance improvements (reduced number of syscalls)
- worker factory connection is now exposed to PHP using RR_RELAY env
- HTTPS support
- HTTP/2 and HTTP/2 Support
- Removed `disable` flag of static service

v1.2.3 (29.09.2018)
------
- reduced verbosity
- worker list has been extracted from http service and now available for other rr based services
- built using Go 1.11

v1.2.2 (23.09.2018)
------
- new project directory structure
- introduces DefaultsConfig, allows to keep config files smaller
- better worker pool destruction while working with long running processes
- added more php versions to travis config
- `Spiral\RoadRunner\Exceptions\RoadRunnerException` is marked as deprecated in favor of `Spiral\RoadRunner\Exception\RoadRunnerException`
- improved test coverage

v1.2.1 (21.09.2018)
------
- added RR_HTTP env variable to php processes run under http service
- bugfix: ignored `--config` option
- added shorthand for config `-c`
- rr now changes working dir to the config location (allows relating paths for php scripts)

v1.2.0 (10.09.2018)
-------
- added an ability to request `*logrus.Logger`, `logrus.StdLogger`, `logrus.FieldLogger` dependency
in container
- added ability to set env values using `env.Environment`
- `env.Provider` renamed to `env.Environment`
- rr does not throw a warning when service config is missing, instead debug level is used
- rr server config now support default value set (shorter configs)
- debug handlers have been moved from root command and now can be defined for each service separately
- bugfix: panic when using debug mode without http service registered
- `rr.Verbose` and `rr.Debug`is not public
- rpc service now exposes it's addressed to underlying workers to simplify the connection
- env service construction has been simplified in order to unify it with other services
- more tests

v1.1.1 (26.07.2018)
-------
- added support for custom env variables
- added env service
- added env provider to provide ability to define env variables from any source
- container can resolve values by interface now

v1.1.0 (08.07.2018)
-------
- bugfix: Wrong values for $_SERVER['REQUEST_TIME'] and $_SERVER['REQUEST_TIME_FLOAT']
- rr now resolves remoteAddr (IP-address)
- improvements in the error buffer
- support for custom configs and dependency injection for services
- support for net/http native middlewares
- better debugger
- config pre-processing now allows seconds for http service timeouts
- support for non-serving services

v1.0.5 (30.06.2018)
-------
- docker compatible logging (forcing TTY output for logrus)

v1.0.4 (25.06.2018)
-------
- changes in server shutdown sequence

v1.0.3 (23.06.2018)
-------
- rr would provide error log from workers in realtime now
- even better service shutdown
- safer unix socket allocation
- minor CS

v1.0.2 (19.06.2018)
-------
- more validations for user configs

v1.0.1 (15.06.2018)
-------
- Makefile added

v1.0.0 (14.06.2018)
------
- higher performance
- worker.State.Updated() has been removed in order to improve overall performance
- staticPool can automatically replace workers killed from outside
- server would not attempt to rebuild static pool in case of reoccurring failure
- PSR-7 server
- file uploads
- service container and plugin based model
- RPC server
- better control over worker state, move events
- static files server
- hot code reload, interactive workers console
- support for future streaming responses
- much higher tests coverage
- less dependencies
- yaml/json configs (thx viper)
- CLI application server
- middleware and event listeners support
- psr7 library for php
