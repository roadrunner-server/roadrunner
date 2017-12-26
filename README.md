RoadRunner
==========
PHP application server library for Golang.

Features:
--------
- load balancer, process manager and task pipeline
- hot-swap of workers
- build for multiple frontends (queue, rest, psr-7, async php, etc)
- works over TPC, unix sockets, standard pipes
- safe worker termination
- timeout management
- payload context
- protocol, job and worker level error management
- very fast (~200k calls per second on Ryzen 1700X over 17 threads)
- works on Windows

License:
--------
The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information.
