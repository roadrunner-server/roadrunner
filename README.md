RoadRunner
==========
[![GoDoc](https://godoc.org/github.com/spiral/roadrunner?status.svg)](https://godoc.org/github.com/spiral/roadrunner)
[![Build Status](https://travis-ci.org/spiral/roadrunner.svg?branch=master)](https://travis-ci.org/spiral/roadrunner)

High-performance PHP application server for Golang.

Features:
--------
- load balancer, process manager and task pipeline 
- swaps workers without stopping the server
- build for multiple frontends (queue, rest, psr-7, async php, etc)
- works over TPC, unix sockets and standard pipes
- automatic worker replacement
- safe worker destruction
- worker lifecycle management (create/stop/allocate timeouts)
- payload context
- protocol, job and worker level error management
- very fast (~200k calls per second on Ryzen 1700X over 17 threads)
- works on Windows

License:
--------
The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information.