<?php

use Spiral\RoadRunner;
use Nyholm\Psr7;

include "vendor/autoload.php";

$worker = RoadRunner\Worker::create();
$psrFactory = new Psr7\Factory\Psr17Factory();

$worker = new RoadRunner\Http\PSR7Worker($worker, $psrFactory, $psrFactory, $psrFactory);
$counter = 0;

while ($req = $worker->waitRequest()) {
    try {
        $rsp = new Psr7\Response();

        if ($req->getUri()->getPath() !== '/') {
            $worker->respond($rsp->withStatus(404));
            continue;
        }

        $rsp->getBody()->write('Hello world!');
        $rsp->getBody()->write(PHP_EOL);
        $rsp->getBody()->write((string)$counter++);

        $worker->respond($rsp);

    } catch (\Throwable $e) {
        $worker->getWorker()->error((string)$e);
    }
}