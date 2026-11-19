<?php

declare(strict_types=1);

use Spiral\RoadRunner\Jobs\Consumer;

require dirname(__DIR__) . '/vendor/autoload.php';

$consumer = new Consumer();
while ($task = $consumer->waitTask()) {
    if (!$task->hasHeader('retried')) {
        $task->withHeader('retried', 'yes')->requeue('retry once');
        continue;
    }

    file_put_contents(getenv('RR_REDIS_JOB_RESULT'), json_encode([
        $task->getId(), $task->getDriver()->value, $task->getPipeline(), $task->getQueue(),
        base64_encode($task->getPayload()), $task->getHeader('tags'), $task->getHeader('retried'),
    ], JSON_THROW_ON_ERROR));
    $task->ack();
}
