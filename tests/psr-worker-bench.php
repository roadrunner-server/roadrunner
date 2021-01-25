<?php

declare(strict_types=1);

use Spiral\RoadRunner;
use Nyholm\Psr7\Factory;

ini_set('display_errors', 'stderr');
include "vendor/autoload.php";

$env = \Spiral\RoadRunner\Environment::fromGlobals();

if ($env->getMode() === 'http') {
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
} else {
    /**
     * @param string $dir
     * @return array<string>
     */
    $getClasses = static function (string $dir): iterable {
        $files = glob($dir . '/*.php');

        foreach ($files as $file) {
            yield substr(basename($file), 0, -4);
        }
    };

    $factory = \Temporal\WorkerFactory::create();

    $worker = $factory->newWorker('default');

    // register all workflows
    foreach ($getClasses(__DIR__ . '/../temporal/Workflow') as $name) {
        $worker->registerWorkflowType('Temporal\\Tests\\Workflow\\' . $name);
    }

    // register all activity
    foreach ($getClasses(__DIR__ . '/../temporal/Activity') as $name) {
        $worker->registerActivityType('Temporal\\Tests\\Activity\\' . $name);
    }

    $factory->run();
}