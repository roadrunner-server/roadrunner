CHANGELOG
=========

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
- middlewares and event listeners
- psr7 library for php
