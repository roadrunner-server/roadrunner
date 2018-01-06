RoadRunner
==========
Embeddable PHP application server library for Golang.

Features:
--------
- load balancer, process manager and task pipeline 
- hot-wrap of worker pool 
- build for multiple frontends (queue, rest, psr-7, async php, etc)
- works over TPC, unix sockets and standard pipes
- controlled worker termination
- timeout management
- payload context
- protocol, job and worker level error management
- very fast (~200k calls per second on Ryzen 1700X over 17 threads)
- works on Windows

License:
--------
The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information.