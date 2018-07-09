CHANGELOG
=========

v1.1.0 (80.07.2018)
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
