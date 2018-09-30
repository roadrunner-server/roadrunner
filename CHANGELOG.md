CHANGELOG
=========

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
