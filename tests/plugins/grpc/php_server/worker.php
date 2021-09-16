<?php
/**
 * Sample GRPC PHP server.
 */

use Service\EchoInterface;
use Spiral\Goridge\StreamRelay;
use Spiral\GRPC\Server;
use Spiral\RoadRunner\Worker;

require __DIR__ . '/vendor/autoload.php';

$server = new Server(null, [
    'debug' => false, // optional (default: false)
]);

$server->registerService(EchoInterface::class, new EchoService());

$worker = \method_exists(Worker::class, 'create')
    // RoadRunner >= 2.x
    ? Worker::create()
    // RoadRunner 1.x
    : new Worker(new StreamRelay(STDIN, STDOUT))
;

$server->serve($worker);
