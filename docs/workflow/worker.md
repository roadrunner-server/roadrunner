# Temporal Worker
Unlike HTTP, Temporal use different way to configure a worker. Make sure to require PHP SDK:

```
$ composer require temporal/sdk
```

The worker file will look as following:

```php
<?php

declare(strict_types=1);

use Temporal\WorkerFactory;

ini_set('display_errors', 'stderr');
include "vendor/autoload.php";

// factory initiates and runs task queue specific activity and workflow workers
$factory = WorkerFactory::create();

// Worker that listens on a task queue and hosts both workflow and activity implementations.
$worker = $factory->newWorker(
    'taskQueue',
    \Temporal\Worker\WorkerOptions::new()->withMaxConcurrentActivityExecutionSize(10)
);

// Workflows are stateful. So you need a type to create instances.
$worker->registerWorkflowTypes(MyWorkflow::class);

// Activities are stateless and thread safe. So a shared instance is used.
$worker->registerActivityImplementations(new MyActivity());


// start primary loop
$factory->run();
```

Read more about temporal configuration and usage [at official website](https://docs.temporal.io/docs/php-sdk-overview). 

## Multi-worker Environment
To serve both HTTP and Temporal from the same worker use `getMode()` option of `Environment`:

```php
use Spiral\RoadRunner;

$rrEnv = RoadRunner\Environment::fromGlobals();

if ($rrEnv->getMode() === RoadRunner\Environment\Mode::MODE_TEMPORAL) {
    // start temporal worker
    return;
}

if ($rrEnv->getMode() === RoadRunner\Environment\Mode::MODE_HTTP) {
    // start http worker
    return;
}
```