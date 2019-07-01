CHANGELOG
=========

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
