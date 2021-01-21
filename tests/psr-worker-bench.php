<?php

use Spiral\RoadRunner;
use Nyholm\Psr7\Factory;

ini_set('display_errors', 'stderr');
include "vendor/autoload.php";

$worker = new RoadRunner\Http\PSR7Worker(
    RoadRunner\Worker::create(),
    new Factory\Psr17Factory(),
    new Factory\Psr17Factory(),
    new Factory\Psr17Factory()
);

while ($req = $worker->waitRequest()) {
    try {
        $rsp = new \Nyholm\Psr7\Response();
        $rsp->getBody()->write("hello world");
        $worker->respond($rsp);
    } catch (\Throwable $e) {
        $worker->getWorker()->error((string)$e);
    }
}