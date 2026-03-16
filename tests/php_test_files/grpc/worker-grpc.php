<?php

/**
 * Sample gRPC PHP server for e2e tests.
 */

use Service\EchoInterface;
use Health\HealthInterface;
use Spiral\RoadRunner\GRPC\Server;
use Spiral\RoadRunner\Worker;

require dirname(__DIR__) . '/vendor/autoload.php';

$server = new Server();

$server->registerService(EchoInterface::class, new EchoService());
$server->registerService(HealthInterface::class, new HealthService());

$server->serve(Worker::create());
