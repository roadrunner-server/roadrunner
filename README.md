RoadRunner [WIP]
==========
[![Latest Stable Version](https://poser.pugx.org/spiral/roadrunner/v/stable)](https://packagist.org/packages/spiral/roadrunner) 
[![GoDoc](https://godoc.org/github.com/spiral/roadrunner?status.svg)](https://godoc.org/github.com/spiral/roadrunner)
[![Build Status](https://travis-ci.org/spiral/roadrunner.svg?branch=master)](https://travis-ci.org/spiral/roadrunner)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/spiral/roadrunner/badges/quality-score.png)](https://scrutinizer-ci.com/g/spiral/roadrunner/?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/spiral/roadrunner)](https://goreportcard.com/report/github.com/spiral/roadrunner)

PHP application server library for Golang.

Features:
--------
- load balancer, process manager and task pipeline
- hot-swap of workers
- build for multiple frontends (queue, rest, psr-7, async php, etc)
- works over TPC, unix sockets, standard pipes
- safe worker termination
- protocol, worker and job level error management
- very fast (~200k calls per second on Ryzen 1700X over 17 threads)
- works on Windows

License:
--------
The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information.
