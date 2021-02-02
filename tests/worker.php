<?php

declare(strict_types=1);

require __DIR__ . '/vendor/autoload.php';

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
foreach ($getClasses(__DIR__ . '/src/Workflow') as $name) {
    $worker->registerWorkflowTypes('Temporal\\Tests\\Workflow\\' . $name);
}

// register all activity
foreach ($getClasses(__DIR__ . '/src/Activity') as $name) {
    $class = 'Temporal\\Tests\\Activity\\' . $name;
    $worker->registerActivityImplementations(new $class);
}

$factory->run();
