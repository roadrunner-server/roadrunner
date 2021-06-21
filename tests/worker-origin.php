<?php

use Spiral\RoadRunner\Worker;
use Spiral\RoadRunner\Http\HttpWorker;

require __DIR__ . '/vendor/autoload.php';

$http = new HttpWorker(Worker::create());

while ($req = $http->waitRequest()) {
    $http->respond(200, 'Response', [
        'Access-Control-Allow-Origin' => ['*']
    ]);
}
