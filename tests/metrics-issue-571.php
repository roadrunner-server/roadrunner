<?php

use Spiral\Goridge;
use Spiral\RoadRunner;
use Nyholm\Psr7\Factory;

ini_set('display_errors', 'stderr');
require __DIR__ . "/vendor/autoload.php";

$worker = new RoadRunner\Http\PSR7Worker(
	RoadRunner\Worker::create(),
	new Factory\Psr17Factory(),
	new Factory\Psr17Factory(),
	new Factory\Psr17Factory()
);

$metrics = new RoadRunner\Metrics\Metrics(
	Goridge\RPC\RPC::create(RoadRunner\Environment::fromGlobals()->getRPCAddress())
);

$metrics->declare(
	'test',
	RoadRunner\Metrics\Collector::counter()->withHelp('Test counter')
);

while ($req = $worker->waitRequest()) {
	try {
		$rsp = new \Nyholm\Psr7\Response();
		$rsp->getBody()->write("hello world");

		$metrics->add('test', 1);

		$worker->respond($rsp);
	} catch (\Throwable $e) {
		$worker->getWorker()->error((string)$e);
	}
}
